// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
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

var agentNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

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

// AgentInstanceResponse is one registered instance of an agent.
type AgentInstanceResponse struct {
	Id         db.Id           `json:"id"`
	InstanceId string          `json:"instanceId"`
	Address    string          `json:"address"`
	Port       int             `json:"port"`
	Datacenter *string         `json:"datacenter,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  string          `json:"createdAt"`
	UpdatedAt  string          `json:"updatedAt"`
}

// AgentResponse is an agent shaped for API responses.
type AgentResponse struct {
	Id           db.Id                    `json:"id"`
	WorkspaceId  db.Id                    `json:"workspaceId"`
	Name         string                   `json:"name"`
	Version      string                   `json:"version"`
	Card         json.RawMessage          `json:"card"`
	CardChecksum string                   `json:"cardChecksum"`
	Health       string                   `json:"health"`
	Instances    []*AgentInstanceResponse `json:"instances"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
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
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

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
		instances, err := a.InstanceRepository.ListByAgentId(item.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
		}

		now := time.Now().UTC()
		statuses := make([]string, 0, len(instances))
		mapped := make([]*AgentInstanceResponse, 0, len(instances))
		for _, instance := range instances {
			checks, err := a.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
			}
			status := instanceStatus(checks, now)
			statuses = append(statuses, status)

			mapped = append(mapped, &AgentInstanceResponse{
				Id:         instance.Id,
				InstanceId: instance.InstanceId,
				Address:    instance.Address,
				Port:       instance.Port,
				Datacenter: instance.Datacenter,
				Meta:       util.JSONRawFromString(instance.Meta),
				Status:     status,
				CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
				UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
			})
		}
		health := worstStatus(statuses)

		list = append(list, &AgentResponse{
			Id:           item.Id,
			WorkspaceId:  item.WorkspaceId,
			Name:         item.Name,
			Version:      item.Version,
			Card:         json.RawMessage(item.Card),
			CardChecksum: item.CardChecksum,
			Health:       health,
			Instances:    mapped,
			CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListAgentsResponse{Agents: list, Total: total}, nil
}

// GetAgent returns one agent by name.
func (a *Agent) GetAgent(workspaceId db.Id, name string) (*AgentResponse, error) {
	if !agentNamePattern.MatchString(name) {
		return nil, ErrInvalidAgentName
	}

	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	instances, err := a.InstanceRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}

	now := time.Now().UTC()
	statuses := make([]string, 0, len(instances))
	mapped := make([]*AgentInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		checks, err := a.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
		}
		status := instanceStatus(checks, now)
		statuses = append(statuses, status)

		mapped = append(mapped, &AgentInstanceResponse{
			Id:         instance.Id,
			InstanceId: instance.InstanceId,
			Address:    instance.Address,
			Port:       instance.Port,
			Datacenter: instance.Datacenter,
			Meta:       util.JSONRawFromString(instance.Meta),
			Status:     status,
			CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	health := worstStatus(statuses)

	return &AgentResponse{
		Id:           agent.Id,
		WorkspaceId:  agent.WorkspaceId,
		Name:         agent.Name,
		Version:      agent.Version,
		Card:         json.RawMessage(agent.Card),
		CardChecksum: agent.CardChecksum,
		Health:       health,
		Instances:    mapped,
		CreatedAt:    agent.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    agent.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// UpsertAgent creates or replaces the agent card for a name.
func (a *Agent) UpsertAgent(workspaceId db.Id, name string, req *UpsertAgentRequest) (*AgentResponse, bool, error) {
	if !agentNamePattern.MatchString(name) {
		return nil, false, ErrInvalidAgentName
	}

	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, false, err
	}
	if workspace == nil {
		return nil, false, ErrWorkspaceNotFound
	}

	if req == nil || len(bytes.TrimSpace(req.Card)) == 0 {
		return nil, false, ErrInvalidAgentCard
	}

	var compact bytes.Buffer
	err = json.Compact(&compact, req.Card)
	if err != nil || compact.Len() == 0 || compact.Bytes()[0] != '{' {
		return nil, false, ErrInvalidAgentCard
	}

	version := req.Version
	if version == "" {
		var body struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(compact.Bytes(), &body) != nil {
			return nil, false, ErrInvalidAgentCard
		}
		version = body.Version
	}
	if version == "" || len(version) > 40 {
		return nil, false, ErrAgentVersionRequired
	}

	sum := sha256.Sum256(compact.Bytes())
	card := compact.String()
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

	instances, err := a.InstanceRepository.ListByAgentId(existing.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	now := time.Now().UTC()
	statuses := make([]string, 0, len(instances))
	mapped := make([]*AgentInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		checks, err := a.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
		if err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
		}
		status := instanceStatus(checks, now)
		statuses = append(statuses, status)

		mapped = append(mapped, &AgentInstanceResponse{
			Id:         instance.Id,
			InstanceId: instance.InstanceId,
			Address:    instance.Address,
			Port:       instance.Port,
			Datacenter: instance.Datacenter,
			Meta:       util.JSONRawFromString(instance.Meta),
			Status:     status,
			CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	health := worstStatus(statuses)

	return &AgentResponse{
		Id:           existing.Id,
		WorkspaceId:  existing.WorkspaceId,
		Name:         existing.Name,
		Version:      existing.Version,
		Card:         json.RawMessage(existing.Card),
		CardChecksum: existing.CardChecksum,
		Health:       health,
		Instances:    mapped,
		CreatedAt:    existing.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    existing.UpdatedAt.UTC().Format(time.RFC3339),
	}, created, nil
}

// DeleteAgent removes an agent by name.
func (a *Agent) DeleteAgent(workspaceId db.Id, name string) error {
	if !agentNamePattern.MatchString(name) {
		return ErrInvalidAgentName
	}

	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

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
