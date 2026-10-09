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
	ErrFailedRenewInstance      = errors.New("failed renew instance")
	ErrFailedDeregisterInstance = errors.New("failed deregister instance")
)

// DefaultLeaseTTL is the lease length in seconds when none is given.
const DefaultLeaseTTL = 30

var instanceIdPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Instance is the module for agent instance registration and leases.
type Instance struct {
	AgentRepository       db.AgentRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	AgentCheckRepository  db.AgentCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
}

// RegisterInstanceRequest is the body for registering or updating an instance.
type RegisterInstanceRequest struct {
	Address    string          `json:"address" validate:"required,max=255" label:"Address"`
	Port       int             `json:"port" validate:"required,min=1,max=65535" label:"Port"`
	Datacenter *string         `json:"datacenter" validate:"omitempty,max=60" label:"Datacenter"`
	Meta       json.RawMessage `json:"meta" label:"Meta"`
	LeaseTTL   int             `json:"leaseTtl" validate:"omitempty,min=5,max=86400" label:"Lease TTL"`
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
	Id             db.Id                  `json:"id"`
	AgentId        db.Id                  `json:"agentId"`
	InstanceId     string                 `json:"instanceId"`
	Address        string                 `json:"address"`
	Port           int                    `json:"port"`
	Datacenter     *string                `json:"datacenter,omitempty"`
	Meta           json.RawMessage        `json:"meta,omitempty"`
	Status         string                 `json:"status"`
	LeaseExpiresAt *time.Time             `json:"leaseExpiresAt,omitempty"`
	Checks         []*HealthCheckResponse `json:"checks"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

// LeaseDefinition is stored as the TTL check definition so renew knows the lease length.
type LeaseDefinition struct {
	TTL int `json:"ttl"`
}

// NewInstance creates an instance module with the given repositories.
func NewInstance(agents db.AgentRepository, instances db.AgentInstanceRepository, checks db.HealthCheckRepository, agentChecks db.AgentCheckRepository, workspaces db.WorkspaceRepository) *Instance {
	return &Instance{
		AgentRepository:       agents,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		AgentCheckRepository:  agentChecks,
		WorkspaceRepository:   workspaces,
	}
}

// ListInstances returns an agent's instances, only live ones when passing is set.
func (i *Instance) ListInstances(workspaceId db.Id, agentName string, passing bool) ([]*InstanceResponse, error) {
	agent, err := i.getAgent(workspaceId, agentName, ErrFailedListInstances)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	var instances []*db.AgentInstance
	if passing {
		instances, err = i.InstanceRepository.ListLiveByAgentId(agent.Id, now)
	} else {
		instances, err = i.InstanceRepository.ListByAgentId(agent.Id)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	list := make([]*InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		item, err := i.toInstanceResponse(instance, now)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}
		list = append(list, item)
	}
	return list, nil
}

// GetInstance returns one instance of an agent.
func (i *Instance) GetInstance(workspaceId db.Id, agentName, instanceId string) (*InstanceResponse, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, ErrInvalidInstanceId
	}

	agent, err := i.getAgent(workspaceId, agentName, ErrFailedGetInstance)
	if err != nil {
		return nil, err
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	item, err := i.toInstanceResponse(instance, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	return item, nil
}

// RegisterInstance creates or updates an instance, starts its lease and syncs its checks.
func (i *Instance) RegisterInstance(workspaceId db.Id, agentName, instanceId string, req *RegisterInstanceRequest) (*InstanceResponse, bool, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, false, ErrInvalidInstanceId
	}

	agent, err := i.getAgent(workspaceId, agentName, ErrFailedRegisterInstance)
	if err != nil {
		return nil, false, err
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
	instance.Datacenter = req.Datacenter
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

	ttl := req.LeaseTTL
	if ttl == 0 {
		ttl = DefaultLeaseTTL
	}
	raw, _ := json.Marshal(LeaseDefinition{TTL: ttl})
	definition := string(raw)
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(ttl) * time.Second)

	lease, err := i.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, db.HealthCheckIDTTL)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}
	if lease == nil {
		lease = &db.HealthCheck{
			AgentInstanceId: instance.Id,
			CheckId:         db.HealthCheckIDTTL,
			Name:            "Lease",
			Type:            db.HealthCheckTypeTTL,
			Source:          db.HealthCheckSourceLease,
			Status:          db.HealthCheckStatusPassing,
			Definition:      &definition,
			TTLExpiresAt:    &expiresAt,
			LastRunAt:       &now,
		}
		err = i.HealthCheckRepository.Create(lease)
	} else {
		lease.Definition = &definition
		err = i.HealthCheckRepository.Update(lease)
		if err == nil {
			err = i.HealthCheckRepository.Pass(lease.Id, "", &expiresAt)
		}
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	templates, err := i.AgentCheckRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}
	err = syncInstanceChecks(i.HealthCheckRepository, instance.Id, templates, now)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	instance.UpdatedAt = now
	item, err := i.toInstanceResponse(instance, now)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}
	return item, created, nil
}

// RenewInstance extends an instance lease by its registered TTL.
func (i *Instance) RenewInstance(workspaceId db.Id, agentName, instanceId string) (*InstanceResponse, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, ErrInvalidInstanceId
	}

	agent, err := i.getAgent(workspaceId, agentName, ErrFailedRenewInstance)
	if err != nil {
		return nil, err
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	// Registering always creates the lease, so it is there to extend.
	lease, err := i.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, db.HealthCheckIDTTL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}
	if lease == nil {
		return nil, ErrInstanceNotFound
	}

	var def LeaseDefinition
	json.Unmarshal([]byte(*lease.Definition), &def)
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(def.TTL) * time.Second)

	err = i.HealthCheckRepository.Pass(lease.Id, "", &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}

	item, err := i.toInstanceResponse(instance, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}
	return item, nil
}

// DeregisterInstance removes an instance and its health checks.
func (i *Instance) DeregisterInstance(workspaceId db.Id, agentName, instanceId string) error {
	if !instanceIdPattern.MatchString(instanceId) {
		return ErrInvalidInstanceId
	}

	agent, err := i.getAgent(workspaceId, agentName, ErrFailedDeregisterInstance)
	if err != nil {
		return err
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

// getAgent loads an agent by name, wrapping lookup failures with failErr.
func (i *Instance) getAgent(workspaceId db.Id, agentName string, failErr error) (*db.Agent, error) {
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
		return nil, fmt.Errorf("%w: %v", failErr, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// toInstanceResponse loads an instance's checks and shapes it for the API.
func (i *Instance) toInstanceResponse(instance *db.AgentInstance, now time.Time) (*InstanceResponse, error) {
	checks, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
	if err != nil {
		return nil, err
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Datacenter: instance.Datacenter,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     instanceStatus(checks, now),
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}
	for _, check := range checks {
		if check.Source == db.HealthCheckSourceLease {
			item.LeaseExpiresAt = check.TTLExpiresAt
		}
		item.Checks = append(item.Checks, toHealthCheckResponse(check, now))
	}
	return item, nil
}
