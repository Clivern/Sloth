// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrInvalidAgentName     = errors.New("invalid agent name")
	ErrInvalidAgentCard     = errors.New("invalid agent card")
	ErrAgentVersionRequired = errors.New("agent version required")
	ErrFailedListAgents     = errors.New("failed list agents")
	ErrFailedGetAgent       = errors.New("failed get agent")
	ErrFailedUpsertAgent    = errors.New("failed upsert agent")
	ErrFailedDeleteAgent    = errors.New("failed delete agent")
)

// AgentNamePattern is the regex pattern for agent names.
var AgentNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Agent is the module for the workspace agent catalog.
type Agent struct {
	AgentRepository       db.AgentRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
}

// UpsertAgentRequest is the body for registering or replacing an agent card.
type UpsertAgentRequest struct {
	Card    json.RawMessage `json:"card" label:"Card"`
	Version string          `json:"version" validate:"omitempty,max=40" label:"Version"`
}

// AgentResponse is an agent shaped for API responses.
type AgentResponse struct {
	Id           db.Id           `json:"id"`
	WorkspaceId  db.Id           `json:"workspaceId"`
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Card         json.RawMessage `json:"card"`
	CardChecksum string          `json:"cardChecksum"`
	Health       string          `json:"health"`
	StatusCounts map[string]int  `json:"statusCounts"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

// ListAgentsResponse is returned when listing agents.
type ListAgentsResponse struct {
	Agents []*AgentResponse
	Total  int64
}

// NewAgent creates an agent module with the given repositories.
func NewAgent(agents db.AgentRepository, instances db.AgentInstanceRepository, checks db.HealthCheckRepository, workspaces db.WorkspaceRepository) *Agent {
	return &Agent{
		AgentRepository:       agents,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		WorkspaceRepository:   workspaces,
	}
}

// ListAgents returns paginated agents in a workspace.
func (a *Agent) ListAgents(workspaceId db.Id, limit, offset int) (*ListAgentsResponse, error) {
	total, err := a.AgentRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
	}

	agents, err := a.AgentRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
	}

	list := make([]*AgentResponse, 0, len(agents))
	for _, item := range agents {
		health, counts, err := a.GetHealthStatus(item.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
		}

		list = append(list, &AgentResponse{
			Id:           item.Id,
			WorkspaceId:  item.WorkspaceId,
			Name:         item.Name,
			Version:      item.Version,
			Card:         json.RawMessage(item.Card),
			CardChecksum: item.CardChecksum,
			Health:       health,
			StatusCounts: counts,
			CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListAgentsResponse{Agents: list, Total: total}, nil
}

// GetAgent returns one agent by name.
func (a *Agent) GetAgent(workspaceId db.Id, name string) (*AgentResponse, error) {
	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	health, counts, err := a.GetHealthStatus(agent.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}

	return &AgentResponse{
		Id:           agent.Id,
		WorkspaceId:  agent.WorkspaceId,
		Name:         agent.Name,
		Version:      agent.Version,
		Card:         json.RawMessage(agent.Card),
		CardChecksum: agent.CardChecksum,
		Health:       health,
		StatusCounts: counts,
		CreatedAt:    agent.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    agent.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// UpsertAgent creates or replaces the agent card for a name.
func (a *Agent) UpsertAgent(workspaceId db.Id, name string, req *UpsertAgentRequest) (*AgentResponse, bool, error) {
	if !AgentNamePattern.MatchString(name) {
		return nil, false, ErrInvalidAgentName
	}

	if req == nil {
		return nil, false, ErrInvalidAgentCard
	}

	var raw map[string]any
	var body struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(req.Card, &raw) != nil || json.Unmarshal(req.Card, &body) != nil {
		return nil, false, ErrInvalidAgentCard
	}

	version := lo.CoalesceOrEmpty(req.Version, body.Version)
	if version == "" || len(version) > 40 {
		return nil, false, ErrAgentVersionRequired
	}

	gateway := fmt.Sprintf(
		"%s/api/v1/workspaces/%s/a2a/agents/%s",
		viper.GetString("app.url"),
		workspaceId,
		name,
	)

	interfaces, _ := raw["supportedInterfaces"].([]any)
	if len(interfaces) == 0 {
		interfaces = []any{map[string]any{}}
	}

	for _, iface := range interfaces {
		iface.(map[string]any)["url"] = gateway
	}

	raw["supportedInterfaces"] = interfaces

	if lo.HasKey(raw, "url") {
		raw["url"] = gateway
	}

	cardJSON, _ := json.Marshal(raw)
	sum := sha256.Sum256(cardJSON)
	card := string(cardJSON)
	checksum := hex.EncodeToString(sum[:])

	existing, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	created := existing == nil
	if created {
		existing = &db.Agent{WorkspaceId: workspaceId, Name: name}
	}

	existing.Name = name
	existing.Card = card
	existing.CardChecksum = checksum
	existing.Version = version

	if created {
		err = a.AgentRepository.Create(existing)
	} else {
		err = a.AgentRepository.Update(existing)
		if err == nil {
			existing, err = a.AgentRepository.GetById(existing.Id)
		}
	}
	if err != nil || existing == nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	health, counts, err := a.GetHealthStatus(existing.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	return &AgentResponse{
		Id:           existing.Id,
		WorkspaceId:  existing.WorkspaceId,
		Name:         existing.Name,
		Version:      existing.Version,
		Card:         json.RawMessage(existing.Card),
		CardChecksum: existing.CardChecksum,
		Health:       health,
		StatusCounts: counts,
		CreatedAt:    existing.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    existing.UpdatedAt.UTC().Format(time.RFC3339),
	}, created, nil
}

// DeleteAgent removes an agent by name.
func (a *Agent) DeleteAgent(workspaceId db.Id, name string) error {
	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteAgent, err)
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	err = a.AgentRepository.Delete(agent.Id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteAgent, err)
	}

	return nil
}

// GetHealthStatus returns the agent health and how many of its instances are in each status.
func (a *Agent) GetHealthStatus(agentId db.Id) (string, map[string]int, error) {
	instances, err := a.InstanceRepository.ListByAgentId(agentId)
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()
	counts := map[string]int{
		db.AgentInstanceStatusPassing:  0,
		db.AgentInstanceStatusWarning:  0,
		db.AgentInstanceStatusUnknown:  0,
		db.AgentInstanceStatusCritical: 0,
	}

	for _, instance := range instances {
		checks, err := a.HealthCheckRepository.ListByInstanceId(instance.Id)
		if err != nil {
			return "", nil, err
		}

		// The instance status is the worst check status. A TTL check past its TTL
		status := db.AgentInstanceStatusPassing
		for _, check := range checks {
			current := lo.Ternary(
				check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)),
				db.HealthCheckStatusCritical,
				check.Status,
			)

			switch current {
			case db.HealthCheckStatusPassing:
			case db.HealthCheckStatusWarning:
				if status == db.AgentInstanceStatusPassing {
					status = db.AgentInstanceStatusWarning
				}
			case db.HealthCheckStatusUnknown:
				if status != db.AgentInstanceStatusCritical {
					status = db.AgentInstanceStatusUnknown
				}
			default:
				status = db.AgentInstanceStatusCritical
			}
		}

		counts[status]++
	}

	// The agent is as healthy as its worst instance. No instances means unknown.
	health := db.AgentInstanceStatusUnknown
	for _, status := range []string{
		db.AgentInstanceStatusCritical,
		db.AgentInstanceStatusUnknown,
		db.AgentInstanceStatusWarning,
		db.AgentInstanceStatusPassing,
	} {
		if counts[status] > 0 {
			health = status
			break
		}
	}

	return health, counts, nil
}
