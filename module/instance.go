// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
)

var (
	ErrInstanceNotFound         = errors.New("instance not found")
	ErrInvalidInstanceId        = errors.New("invalid instance id")
	ErrFailedListInstances      = errors.New("failed list instances")
	ErrFailedGetInstance        = errors.New("failed get instance")
	ErrFailedRegisterInstance   = errors.New("failed register instance")
	ErrFailedDeregisterInstance = errors.New("failed deregister instance")
)

var instanceIdPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Instance is the module for agent instance registration and health.
type Instance struct {
	AgentRepository       db.AgentRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	AgentCheckRepository  db.AgentCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
}

// RegisterInstanceRequest is the body for registering or updating an instance.
type RegisterInstanceRequest struct {
	Address string          `json:"address" validate:"required,max=255" label:"Address"`
	Port    int             `json:"port" validate:"required,min=1,max=65535" label:"Port"`
	Meta    json.RawMessage `json:"meta" label:"Meta"`
}

// HealthCheckResponse is one health check on an instance.
type HealthCheckResponse struct {
	CheckId      string     `json:"checkId"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	Output       *string    `json:"output,omitempty"`
	TTLExpiresAt *time.Time `json:"ttlExpiresAt,omitempty"`
	LastRunAt    *time.Time `json:"lastRunAt,omitempty"`
}

// InstanceResponse is an instance shaped for API responses.
type InstanceResponse struct {
	Id         db.Id                  `json:"id"`
	AgentId    db.Id                  `json:"agentId"`
	InstanceId string                 `json:"instanceId"`
	Address    string                 `json:"address"`
	Port       int                    `json:"port"`
	Meta       json.RawMessage        `json:"meta,omitempty"`
	Status     string                 `json:"status"`
	Checks     []*HealthCheckResponse `json:"checks"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// ListInstancesResponse is one page of an agent's instances and the total across all pages.
type ListInstancesResponse struct {
	Instances []*InstanceResponse
	Total     int64
}

// NewInstance creates an instance module with the given repositories.
func NewInstance(
	agents db.AgentRepository,
	instances db.AgentInstanceRepository,
	checks db.HealthCheckRepository,
	agentChecks db.AgentCheckRepository,
	workspaces db.WorkspaceRepository,
) *Instance {
	return &Instance{
		AgentRepository:       agents,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		AgentCheckRepository:  agentChecks,
		WorkspaceRepository:   workspaces,
	}
}

// ListInstances returns one page of an agent's instances, only live ones when passing is set.
func (i *Instance) ListInstances(
	workspaceId db.Id,
	agentName string,
	passing bool,
	limit, offset int,
) (*ListInstancesResponse, error) {
	if !agentNamePattern.MatchString(agentName) {
		return nil, ErrInvalidAgentName
	}

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	now := time.Now().UTC()

	var instances []*db.AgentInstance
	var total int64
	if passing {
		total, err = i.InstanceRepository.CountLiveByAgentId(agent.Id, now)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		instances, err = i.InstanceRepository.ListLivePageByAgentId(agent.Id, now, limit, offset)
	} else {
		total, err = i.InstanceRepository.CountByAgentId(agent.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		instances, err = i.InstanceRepository.ListPageByAgentId(agent.Id, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	list := make([]*InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		checks, err := i.HealthCheckRepository.ListByInstanceId(instance.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		item := &InstanceResponse{
			Id:         instance.Id,
			AgentId:    instance.AgentId,
			InstanceId: instance.InstanceId,
			Address:    instance.Address,
			Port:       instance.Port,
			Meta:       util.JSONRawFromString(instance.Meta),
			Status:     db.AgentInstanceStatusPassing,
			Checks:     make([]*HealthCheckResponse, 0, len(checks)),
			CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
		}

		// The instance status is the worst check status. A TTL check past its TTL
		for _, check := range checks {
			status := check.Status
			if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
				status = db.HealthCheckStatusCritical
			}

			switch status {
			case db.HealthCheckStatusPassing:
			case db.HealthCheckStatusWarning:
				if item.Status == db.AgentInstanceStatusPassing {
					item.Status = db.AgentInstanceStatusWarning
				}
			default:
				item.Status = db.AgentInstanceStatusCritical
			}

			item.Checks = append(item.Checks, &HealthCheckResponse{
				CheckId:      check.CheckId,
				Name:         check.Name,
				Type:         check.Type,
				Source:       check.Source,
				Status:       status,
				Output:       check.Output,
				TTLExpiresAt: check.TTLExpiresAt,
				LastRunAt:    check.LastRunAt,
			})
		}

		list = append(list, item)
	}

	return &ListInstancesResponse{Instances: list, Total: total}, nil
}

// GetInstance returns one instance of an agent.
func (i *Instance) GetInstance(
	workspaceId db.Id,
	agentName, instanceId string,
) (*InstanceResponse, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, ErrInvalidInstanceId
	}

	if !agentNamePattern.MatchString(agentName) {
		return nil, ErrInvalidAgentName
	}

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	now := time.Now().UTC()
	checks, err := i.HealthCheckRepository.ListByInstanceId(instance.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     db.AgentInstanceStatusPassing,
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// The instance status is the worst check status. A TTL check past its TTL
	for _, check := range checks {
		status := check.Status
		if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
			status = db.HealthCheckStatusCritical
		}

		switch status {
		case db.HealthCheckStatusPassing:
		case db.HealthCheckStatusWarning:
			if item.Status == db.AgentInstanceStatusPassing {
				item.Status = db.AgentInstanceStatusWarning
			}
		default:
			item.Status = db.AgentInstanceStatusCritical
		}

		item.Checks = append(item.Checks, &HealthCheckResponse{
			CheckId:      check.CheckId,
			Name:         check.Name,
			Type:         check.Type,
			Source:       check.Source,
			Status:       status,
			Output:       check.Output,
			TTLExpiresAt: check.TTLExpiresAt,
			LastRunAt:    check.LastRunAt,
		})
	}

	return item, nil
}

// RegisterInstance creates or updates an instance and syncs its checks with the agent's templates.
func (i *Instance) RegisterInstance(
	workspaceId db.Id,
	agentName, instanceId string,
	req *RegisterInstanceRequest,
) (*InstanceResponse, bool, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, false, ErrInvalidInstanceId
	}

	if !agentNamePattern.MatchString(agentName) {
		return nil, false, ErrInvalidAgentName
	}

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, false, err
	}
	if workspace == nil {
		return nil, false, ErrWorkspaceNotFound
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}
	if agent == nil {
		return nil, false, ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	created := instance == nil
	if created {
		instance = &db.AgentInstance{AgentId: agent.Id, InstanceId: instanceId}
	}

	instance.Address = req.Address
	instance.Port = req.Port
	instance.Meta = nil
	if len(req.Meta) > 0 {
		meta := string(req.Meta)
		instance.Meta = &meta
	}

	instance.Status = db.AgentInstanceStatusPassing

	if created {
		err = i.InstanceRepository.Create(instance)
	} else {
		err = i.InstanceRepository.Update(instance)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	now := time.Now().UTC()

	templates, err := i.AgentCheckRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	// Make the instance's checks match the agent's templates. Pull checks run
	// right away and a TTL check gets one TTL window for its first report.
	wanted := make(map[string]bool, len(templates))
	for _, template := range templates {
		wanted[template.CheckId] = true

		existing, err := i.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, template.CheckId)
		if err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
		}

		// A type change resets the check, since its old status no longer means anything.
		if existing != nil && existing.Type != template.Type {
			err = i.HealthCheckRepository.Delete(existing.Id)
			if err != nil {
				return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
			}

			existing = nil
		}

		isPull := template.Type == db.HealthCheckTypeHTTP || template.Type == db.HealthCheckTypeTCP
		stored := template.Definition
		if existing != nil {
			existing.Name = template.Name
			existing.Definition = &stored
			existing.NextRunAt = nil
			if isPull {
				existing.NextRunAt = &now
			}

			err = i.HealthCheckRepository.Update(existing)
		} else {
			check := &db.HealthCheck{
				InstanceId: instance.Id,
				CheckId:    template.CheckId,
				Name:       template.Name,
				Type:       template.Type,
				Source:     db.HealthCheckSourceAgent,
				Status:     db.HealthCheckStatusCritical,
				Definition: &stored,
			}
			if isPull {
				output := "Waiting for the first check"
				check.Output = &output
				check.NextRunAt = &now
			} else {
				var templateDefinition CheckDefinition
				_ = json.Unmarshal([]byte(template.Definition), &templateDefinition)
				output := "Waiting for the first report"
				expiresAt := now.Add(time.Duration(templateDefinition.TTL) * time.Second)
				check.Status = db.HealthCheckStatusPassing
				check.Output = &output
				check.TTLExpiresAt = &expiresAt
			}

			err = i.HealthCheckRepository.Create(check)
		}
		if err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
		}
	}

	// Remove copies of templates the agent no longer has.
	current, err := i.HealthCheckRepository.ListByInstanceId(instance.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	for _, check := range current {
		if check.Source == db.HealthCheckSourceAgent && !wanted[check.CheckId] {
			err = i.HealthCheckRepository.Delete(check.Id)
			if err != nil {
				return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
			}
		}
	}

	instance.UpdatedAt = now
	checks, err := i.HealthCheckRepository.ListByInstanceId(instance.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     db.AgentInstanceStatusPassing,
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// The instance status is the worst check status. A TTL check past its TTL
	for _, check := range checks {
		status := check.Status
		if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
			status = db.HealthCheckStatusCritical
		}

		switch status {
		case db.HealthCheckStatusPassing:
		case db.HealthCheckStatusWarning:
			if item.Status == db.AgentInstanceStatusPassing {
				item.Status = db.AgentInstanceStatusWarning
			}
		default:
			item.Status = db.AgentInstanceStatusCritical
		}

		item.Checks = append(item.Checks, &HealthCheckResponse{
			CheckId:      check.CheckId,
			Name:         check.Name,
			Type:         check.Type,
			Source:       check.Source,
			Status:       status,
			Output:       check.Output,
			TTLExpiresAt: check.TTLExpiresAt,
			LastRunAt:    check.LastRunAt,
		})
	}

	return item, created, nil
}

// DeregisterInstance removes an instance and its health checks.
func (i *Instance) DeregisterInstance(workspaceId db.Id, agentName, instanceId string) error {
	if !instanceIdPattern.MatchString(instanceId) {
		return ErrInvalidInstanceId
	}

	if !agentNamePattern.MatchString(agentName) {
		return ErrInvalidAgentName
	}

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}
	if instance == nil {
		return ErrInstanceNotFound
	}

	err = i.InstanceRepository.Delete(instance.Id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}

	return nil
}
