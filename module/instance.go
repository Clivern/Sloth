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
		checks, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		item := &InstanceResponse{
			Id:         instance.Id,
			AgentId:    instance.AgentId,
			InstanceId: instance.InstanceId,
			Address:    instance.Address,
			Port:       instance.Port,
			Datacenter: instance.Datacenter,
			Meta:       util.JSONRawFromString(instance.Meta),
			Status:     db.AgentInstanceStatusPassing,
			Checks:     make([]*HealthCheckResponse, 0, len(checks)),
			CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
		}

		// The instance status is the worst check status. A TTL check past its TTL
		// is critical, and so is an instance without a lease.
		hasLease := false
		for _, check := range checks {
			status := check.Status
			if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
				status = db.HealthCheckStatusCritical
			}
			if check.Source == db.HealthCheckSourceLease {
				hasLease = true
				item.LeaseExpiresAt = check.TTLExpiresAt
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
		if !hasLease {
			item.Status = db.AgentInstanceStatusCritical
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
	checks, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Datacenter: instance.Datacenter,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     db.AgentInstanceStatusPassing,
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// The instance status is the worst check status. A TTL check past its TTL
	// is critical, and so is an instance without a lease.
	hasLease := false
	for _, check := range checks {
		status := check.Status
		if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
			status = db.HealthCheckStatusCritical
		}
		if check.Source == db.HealthCheckSourceLease {
			hasLease = true
			item.LeaseExpiresAt = check.TTLExpiresAt
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
	if !hasLease {
		item.Status = db.AgentInstanceStatusCritical
	}
	return item, nil
}

// RegisterInstance creates or updates an instance, starts its lease and syncs its checks.
func (i *Instance) RegisterInstance(workspaceId db.Id, agentName, instanceId string, req *RegisterInstanceRequest) (*InstanceResponse, bool, error) {
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

	// Make the instance's checks match the agent's templates. Pull checks run
	// right away and a TTL check gets one TTL window for its first report.
	wanted := make(map[string]bool, len(templates))
	for _, template := range templates {
		wanted[template.CheckId] = true

		existing, err := i.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, template.CheckId)
		if err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
		}
		if existing != nil && existing.Source == db.HealthCheckSourceLease {
			continue
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
				AgentInstanceId: instance.Id,
				CheckId:         template.CheckId,
				Name:            template.Name,
				Type:            template.Type,
				Source:          db.HealthCheckSourceAgent,
				Status:          db.HealthCheckStatusCritical,
				Definition:      &stored,
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
	current, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
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
	checks, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Datacenter: instance.Datacenter,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     db.AgentInstanceStatusPassing,
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// The instance status is the worst check status. A TTL check past its TTL
	// is critical, and so is an instance without a lease.
	hasLease := false
	for _, check := range checks {
		status := check.Status
		if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
			status = db.HealthCheckStatusCritical
		}
		if check.Source == db.HealthCheckSourceLease {
			hasLease = true
			item.LeaseExpiresAt = check.TTLExpiresAt
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
	if !hasLease {
		item.Status = db.AgentInstanceStatusCritical
	}
	return item, created, nil
}

// RenewInstance extends an instance lease by its registered TTL.
func (i *Instance) RenewInstance(workspaceId db.Id, agentName, instanceId string) (*InstanceResponse, error) {
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
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
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

	checks, err := i.HealthCheckRepository.ListByAgentInstanceId(instance.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedRenewInstance, err)
	}

	item := &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Datacenter: instance.Datacenter,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     db.AgentInstanceStatusPassing,
		Checks:     make([]*HealthCheckResponse, 0, len(checks)),
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// The instance status is the worst check status. A TTL check past its TTL
	// is critical, and so is an instance without a lease.
	hasLease := false
	for _, check := range checks {
		status := check.Status
		if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
			status = db.HealthCheckStatusCritical
		}
		if check.Source == db.HealthCheckSourceLease {
			hasLease = true
			item.LeaseExpiresAt = check.TTLExpiresAt
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
	if !hasLease {
		item.Status = db.AgentInstanceStatusCritical
	}
	return item, nil
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
