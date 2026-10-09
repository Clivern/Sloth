// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/cognit/db"
)

var (
	ErrCheckNotFound        = errors.New("check not found")
	ErrInvalidCheckId       = errors.New("invalid check id")
	ErrReservedCheckId      = errors.New("reserved check id")
	ErrUnsupportedCheckType = errors.New("unsupported check type")
	ErrInvalidCheckPath     = errors.New("invalid check path")
	ErrInvalidCheckScheme   = errors.New("invalid check scheme")
	ErrInvalidCheckTiming   = errors.New("invalid check timing")
	ErrInvalidCheckStatus   = errors.New("invalid check status")
	ErrCheckNotReportable   = errors.New("check not reportable")
	ErrFailedListChecks     = errors.New("failed list checks")
	ErrFailedUpsertCheck    = errors.New("failed upsert check")
	ErrFailedDeleteCheck    = errors.New("failed delete check")
	ErrFailedReportCheck    = errors.New("failed report check")
)

const (
	// DefaultCheckInterval is how often http and tcp checks run, in seconds.
	DefaultCheckInterval = 10
	// DefaultCheckTimeout is how long an http or tcp probe may take, in seconds.
	DefaultCheckTimeout = 2
	// MaxCheckOutput caps the output stored for a check.
	MaxCheckOutput = 4096
)

// CheckDefinition is the stored definition of a check template and its instance copies.
type CheckDefinition struct {
	Path     string `json:"path,omitempty"`
	Scheme   string `json:"scheme,omitempty"`
	Port     int    `json:"port,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

// UpsertAgentCheckRequest is the body for creating or replacing an agent check template.
type UpsertAgentCheckRequest struct {
	Name     string `json:"name" validate:"omitempty,max=120" label:"Name"`
	Type     string `json:"type" validate:"required,oneof=ttl http tcp" label:"Type"`
	Path     string `json:"path" validate:"omitempty,max=255" label:"Path"`
	Scheme   string `json:"scheme" validate:"omitempty,oneof=http https" label:"Scheme"`
	Port     int    `json:"port" validate:"omitempty,min=1,max=65535" label:"Port"`
	Interval int    `json:"interval" validate:"omitempty,min=5,max=3600" label:"Interval"`
	Timeout  int    `json:"timeout" validate:"omitempty,min=1,max=60" label:"Timeout"`
	TTL      int    `json:"ttl" validate:"omitempty,min=5,max=86400" label:"TTL"`
}

// ReportCheckRequest is the body an instance sends to update one of its TTL checks.
type ReportCheckRequest struct {
	Output string `json:"output" validate:"omitempty,max=4096" label:"Output"`
}

// AgentCheckResponse is an agent check template shaped for API responses.
type AgentCheckResponse struct {
	CheckId   string `json:"checkId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Path      string `json:"path,omitempty"`
	Scheme    string `json:"scheme,omitempty"`
	Port      int    `json:"port,omitempty"`
	Interval  int    `json:"interval,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Check is the module for agent check templates and instance check reports.
type Check struct {
	AgentRepository       db.AgentRepository
	AgentCheckRepository  db.AgentCheckRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
}

// NewCheck creates a check module with the given repositories.
func NewCheck(agents db.AgentRepository, agentChecks db.AgentCheckRepository, instances db.AgentInstanceRepository, checks db.HealthCheckRepository, workspaces db.WorkspaceRepository) *Check {
	return &Check{
		AgentRepository:       agents,
		AgentCheckRepository:  agentChecks,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		WorkspaceRepository:   workspaces,
	}
}

// ListAgentChecks returns the check templates of an agent.
func (c *Check) ListAgentChecks(workspaceId db.Id, agentName string) ([]*AgentCheckResponse, error) {
	agent, err := c.getAgent(workspaceId, agentName, ErrFailedListChecks)
	if err != nil {
		return nil, err
	}

	templates, err := c.AgentCheckRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListChecks, err)
	}

	list := make([]*AgentCheckResponse, 0, len(templates))
	for _, template := range templates {
		list = append(list, toAgentCheckResponse(template))
	}
	return list, nil
}

// UpsertAgentCheck creates or replaces a check template and applies it to every instance.
func (c *Check) UpsertAgentCheck(workspaceId db.Id, agentName, checkId string, req *UpsertAgentCheckRequest) (*AgentCheckResponse, bool, error) {
	if !instanceIdPattern.MatchString(checkId) {
		return nil, false, ErrInvalidCheckId
	}
	if checkId == db.HealthCheckIDTTL {
		return nil, false, ErrReservedCheckId
	}

	definition, err := normalizeCheck(req)
	if err != nil {
		return nil, false, err
	}
	raw, _ := json.Marshal(definition)

	agent, err := c.getAgent(workspaceId, agentName, ErrFailedUpsertCheck)
	if err != nil {
		return nil, false, err
	}

	template, err := c.AgentCheckRepository.GetByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertCheck, err)
	}

	created := template == nil
	if created {
		template = &db.AgentCheck{AgentId: agent.Id, CheckId: checkId}
	}
	template.Name = strings.TrimSpace(req.Name)
	if template.Name == "" {
		template.Name = checkId
	}
	template.Type = req.Type
	template.Definition = string(raw)

	if created {
		err = c.AgentCheckRepository.Create(template)
	} else {
		err = c.AgentCheckRepository.Update(template)
	}
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertCheck, err)
	}

	instances, err := c.InstanceRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertCheck, err)
	}
	now := time.Now().UTC()
	for _, instance := range instances {
		err = applyCheckTemplate(c.HealthCheckRepository, instance.Id, template, now)
		if err != nil {
			return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertCheck, err)
		}
	}

	return toAgentCheckResponse(template), created, nil
}

// DeleteAgentCheck removes an agent check template and its copies on every instance.
func (c *Check) DeleteAgentCheck(workspaceId db.Id, agentName, checkId string) error {
	if !instanceIdPattern.MatchString(checkId) {
		return ErrInvalidCheckId
	}

	agent, err := c.getAgent(workspaceId, agentName, ErrFailedDeleteCheck)
	if err != nil {
		return err
	}

	template, err := c.AgentCheckRepository.GetByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteCheck, err)
	}
	if template == nil {
		return ErrCheckNotFound
	}

	err = c.HealthCheckRepository.DeleteByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteCheck, err)
	}
	err = c.AgentCheckRepository.Delete(template.Id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteCheck, err)
	}
	return nil
}

// ReportCheck records an instance's report for one of its TTL checks and restarts the TTL.
func (c *Check) ReportCheck(workspaceId db.Id, agentName, instanceId, checkId, status string, req *ReportCheckRequest) (*HealthCheckResponse, error) {
	if !instanceIdPattern.MatchString(instanceId) {
		return nil, ErrInvalidInstanceId
	}
	if !instanceIdPattern.MatchString(checkId) {
		return nil, ErrInvalidCheckId
	}
	if status != db.HealthCheckStatusPassing && status != db.HealthCheckStatusWarning && status != db.HealthCheckStatusCritical {
		return nil, ErrInvalidCheckStatus
	}

	agent, err := c.getAgent(workspaceId, agentName, ErrFailedReportCheck)
	if err != nil {
		return nil, err
	}

	instance, err := c.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedReportCheck, err)
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	check, err := c.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, checkId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedReportCheck, err)
	}
	if check == nil {
		return nil, ErrCheckNotFound
	}
	if check.Type != db.HealthCheckTypeTTL || check.Source == db.HealthCheckSourceLease {
		return nil, ErrCheckNotReportable
	}

	definition := parseCheckDefinition(check.Definition)
	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(definition.TTL) * time.Second)
	output := truncateOutput(req.Output)

	err = c.HealthCheckRepository.Report(check.Id, status, output, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedReportCheck, err)
	}

	check.Status = status
	check.TTLExpiresAt = &expiresAt
	check.LastRunAt = &now
	check.Output = nil
	if output != "" {
		check.Output = &output
	}
	return toHealthCheckResponse(check, now), nil
}

// getAgent loads an agent by name, wrapping lookup failures with failErr.
func (c *Check) getAgent(workspaceId db.Id, agentName string, failErr error) (*db.Agent, error) {
	if !agentNamePattern.MatchString(agentName) {
		return nil, ErrInvalidAgentName
	}

	workspace, err := c.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	agent, err := c.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", failErr, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// normalizeCheck validates a template request and fills in defaults for its type.
func normalizeCheck(req *UpsertAgentCheckRequest) (*CheckDefinition, error) {
	switch req.Type {
	case db.HealthCheckTypeTTL:
		if req.TTL == 0 {
			return nil, ErrInvalidCheckTiming
		}
		return &CheckDefinition{TTL: req.TTL}, nil

	case db.HealthCheckTypeHTTP, db.HealthCheckTypeTCP:
		definition := &CheckDefinition{
			Port:     req.Port,
			Interval: req.Interval,
			Timeout:  req.Timeout,
		}
		if definition.Interval == 0 {
			definition.Interval = DefaultCheckInterval
		}
		if definition.Timeout == 0 {
			definition.Timeout = DefaultCheckTimeout
		}
		if definition.Timeout >= definition.Interval {
			return nil, ErrInvalidCheckTiming
		}

		if req.Type == db.HealthCheckTypeHTTP {
			if !strings.HasPrefix(req.Path, "/") || strings.ContainsAny(req.Path, " \t\r\n") {
				return nil, ErrInvalidCheckPath
			}
			definition.Path = req.Path
			definition.Scheme = req.Scheme
			if definition.Scheme == "" {
				definition.Scheme = "http"
			}
			if definition.Scheme != "http" && definition.Scheme != "https" {
				return nil, ErrInvalidCheckScheme
			}
		}
		return definition, nil

	default:
		return nil, ErrUnsupportedCheckType
	}
}

// applyCheckTemplate creates or updates the copy of a template on one instance.
func applyCheckTemplate(checks db.HealthCheckRepository, instanceId db.Id, template *db.AgentCheck, now time.Time) error {
	existing, err := checks.GetByInstanceAndCheckId(instanceId, template.CheckId)
	if err != nil {
		return err
	}
	if existing != nil && existing.Source == db.HealthCheckSourceLease {
		return nil
	}

	// A type change resets the check, since its old status no longer means anything.
	if existing != nil && existing.Type != template.Type {
		err = checks.Delete(existing.Id)
		if err != nil {
			return err
		}
		existing = nil
	}

	definition := template.Definition
	if existing != nil {
		existing.Name = template.Name
		existing.Definition = &definition
		existing.NextRunAt = nil
		if isPullCheck(template.Type) {
			existing.NextRunAt = &now
		}
		return checks.Update(existing)
	}

	check := &db.HealthCheck{
		AgentInstanceId: instanceId,
		CheckId:         template.CheckId,
		Name:            template.Name,
		Type:            template.Type,
		Source:          db.HealthCheckSourceAgent,
		Status:          db.HealthCheckStatusCritical,
		Definition:      &definition,
	}
	if isPullCheck(template.Type) {
		output := "Waiting for the first check"
		check.Output = &output
		check.NextRunAt = &now
	} else {
		output := "Waiting for the first report"
		expiresAt := now.Add(time.Duration(parseCheckDefinition(&definition).TTL) * time.Second)
		check.Status = db.HealthCheckStatusPassing
		check.Output = &output
		check.TTLExpiresAt = &expiresAt
	}
	return checks.Create(check)
}

// syncInstanceChecks makes an instance's checks match the agent's templates.
func syncInstanceChecks(checks db.HealthCheckRepository, instanceId db.Id, templates []*db.AgentCheck, now time.Time) error {
	wanted := make(map[string]bool, len(templates))
	for _, template := range templates {
		wanted[template.CheckId] = true
		err := applyCheckTemplate(checks, instanceId, template, now)
		if err != nil {
			return err
		}
	}

	current, err := checks.ListByAgentInstanceId(instanceId)
	if err != nil {
		return err
	}
	for _, check := range current {
		if check.Source == db.HealthCheckSourceAgent && !wanted[check.CheckId] {
			err = checks.Delete(check.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isPullCheck(checkType string) bool {
	return checkType == db.HealthCheckTypeHTTP || checkType == db.HealthCheckTypeTCP
}

func parseCheckDefinition(raw *string) *CheckDefinition {
	definition := &CheckDefinition{}
	if raw != nil {
		_ = json.Unmarshal([]byte(*raw), definition)
	}
	return definition
}

func truncateOutput(output string) string {
	if len(output) > MaxCheckOutput {
		return output[:MaxCheckOutput]
	}
	return output
}

// effectiveCheckStatus is a check's status, or critical once its TTL has run out.
func effectiveCheckStatus(check *db.HealthCheck, now time.Time) string {
	if check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)) {
		return db.HealthCheckStatusCritical
	}
	return check.Status
}

func statusRank(status string) int {
	switch status {
	case db.HealthCheckStatusPassing:
		return 0
	case db.HealthCheckStatusWarning:
		return 1
	default:
		return 2
	}
}

// instanceStatus is the worst status across an instance's checks, or critical without a lease.
func instanceStatus(checks []*db.HealthCheck, now time.Time) string {
	hasLease := false
	status := db.AgentInstanceStatusPassing
	for _, check := range checks {
		if check.Source == db.HealthCheckSourceLease {
			hasLease = true
		}
		current := effectiveCheckStatus(check, now)
		if statusRank(current) > statusRank(status) {
			status = current
		}
	}
	if !hasLease {
		return db.AgentInstanceStatusCritical
	}
	return status
}

// worstStatus folds instance statuses into an agent's health, or critical without instances.
func worstStatus(statuses []string) string {
	if len(statuses) == 0 {
		return db.AgentInstanceStatusCritical
	}
	status := db.AgentInstanceStatusPassing
	for _, current := range statuses {
		if statusRank(current) > statusRank(status) {
			status = current
		}
	}
	return status
}

func toHealthCheckResponse(check *db.HealthCheck, now time.Time) *HealthCheckResponse {
	return &HealthCheckResponse{
		CheckId:      check.CheckId,
		Name:         check.Name,
		Type:         check.Type,
		Source:       check.Source,
		Status:       effectiveCheckStatus(check, now),
		Output:       check.Output,
		TTLExpiresAt: check.TTLExpiresAt,
		LastRunAt:    check.LastRunAt,
	}
}

func toAgentCheckResponse(template *db.AgentCheck) *AgentCheckResponse {
	definition := parseCheckDefinition(&template.Definition)
	return &AgentCheckResponse{
		CheckId:   template.CheckId,
		Name:      template.Name,
		Type:      template.Type,
		Path:      definition.Path,
		Scheme:    definition.Scheme,
		Port:      definition.Port,
		Interval:  definition.Interval,
		Timeout:   definition.Timeout,
		TTL:       definition.TTL,
		CreatedAt: template.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: template.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
