// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/cognit/db"
)

var (
	ErrIntentionNotFound      = errors.New("intention not found")
	ErrInvalidIntentionAction = errors.New("invalid intention action")
	ErrFailedListIntentions   = errors.New("failed list intentions")
	ErrFailedGetIntention     = errors.New("failed get intention")
	ErrFailedCreateIntention  = errors.New("failed create intention")
	ErrFailedUpdateIntention  = errors.New("failed update intention")
	ErrFailedDeleteIntention  = errors.New("failed delete intention")
	ErrFailedCheckIntention   = errors.New("failed check intention")
)

// Intention is the module for workspace call policy.
type Intention struct {
	IntentionRepository db.IntentionRepository
	WorkspaceRepository db.WorkspaceRepository
}

// IntentionSource names the calling agent.
type IntentionSource struct {
	Agent string `json:"agent" validate:"required,max=100" label:"Source agent"`
}

// IntentionDestination names the called agent and an optional skill.
type IntentionDestination struct {
	Agent string `json:"agent" validate:"required,max=100" label:"Destination agent"`
	Skill string `json:"skill" validate:"omitempty,max=120" label:"Destination skill"`
}

// SaveIntentionRequest is the body for creating or updating an intention.
type SaveIntentionRequest struct {
	Source      IntentionSource      `json:"source" label:"Source"`
	Destination IntentionDestination `json:"destination" label:"Destination"`
	Action      string               `json:"action" validate:"required,oneof=allow deny" label:"Action"`
}

// CheckIntentionRequest is the body for an intention check.
type CheckIntentionRequest struct {
	Source      IntentionSource      `json:"source" label:"Source"`
	Destination IntentionDestination `json:"destination" label:"Destination"`
}

// IntentionResponse is an intention shaped for API responses.
type IntentionResponse struct {
	Id          db.Id  `json:"id"`
	WorkspaceId db.Id  `json:"workspaceId"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Skill       string `json:"skill,omitempty"`
	Action      string `json:"action"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ListIntentionsResponse is returned when listing intentions.
type ListIntentionsResponse struct {
	Intentions []*IntentionResponse
	Total      int64
}

// CheckIntentionResponse is the decision for one call.
type CheckIntentionResponse struct {
	Action    string             `json:"action"`
	Intention *IntentionResponse `json:"intention,omitempty"`
}

// NewIntention creates an intention module with the given repositories.
func NewIntention(intentions db.IntentionRepository, workspaces db.WorkspaceRepository) *Intention {
	return &Intention{
		IntentionRepository: intentions,
		WorkspaceRepository: workspaces,
	}
}

// ListIntentions returns paginated intentions in a workspace.
func (i *Intention) ListIntentions(workspaceId db.Id, limit, offset int) (*ListIntentionsResponse, error) {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	total, err := i.IntentionRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntentions, err)
	}

	intentions, err := i.IntentionRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntentions, err)
	}

	list := make([]*IntentionResponse, 0, len(intentions))
	for _, item := range intentions {
		skill := ""
		if item.Skill != nil {
			skill = *item.Skill
		}

		list = append(list, &IntentionResponse{
			Id:          item.Id,
			WorkspaceId: item.WorkspaceId,
			Source:      item.SourceAgent,
			Destination: item.DestinationAgent,
			Skill:       skill,
			Action:      item.Action,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListIntentionsResponse{Intentions: list, Total: total}, nil
}

// GetIntention returns one intention by id.
func (i *Intention) GetIntention(workspaceId, intentionId db.Id) (*IntentionResponse, error) {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return nil, ErrIntentionNotFound
	}

	skill := ""
	if intention.Skill != nil {
		skill = *intention.Skill
	}

	return &IntentionResponse{
		Id:          intention.Id,
		WorkspaceId: intention.WorkspaceId,
		Source:      intention.SourceAgent,
		Destination: intention.DestinationAgent,
		Skill:       skill,
		Action:      intention.Action,
		CreatedAt:   intention.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   intention.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// CreateIntention inserts an allow or deny rule.
func (i *Intention) CreateIntention(workspaceId db.Id, req *SaveIntentionRequest) (*IntentionResponse, error) {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}
	if req.Action != db.IntentionActionAllow && req.Action != db.IntentionActionDeny {
		return nil, ErrInvalidIntentionAction
	}

	var skill *string
	if req.Destination.Skill != "" {
		value := req.Destination.Skill
		skill = &value
	}

	intention := &db.Intention{
		WorkspaceId:      workspaceId,
		SourceAgent:      req.Source.Agent,
		DestinationAgent: req.Destination.Agent,
		Skill:            skill,
		Action:           req.Action,
	}
	err = i.IntentionRepository.Create(intention)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateIntention, err)
	}

	storedSkill := ""
	if intention.Skill != nil {
		storedSkill = *intention.Skill
	}

	return &IntentionResponse{
		Id:          intention.Id,
		WorkspaceId: intention.WorkspaceId,
		Source:      intention.SourceAgent,
		Destination: intention.DestinationAgent,
		Skill:       storedSkill,
		Action:      intention.Action,
		CreatedAt:   intention.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   intention.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// UpdateIntention replaces an existing rule.
func (i *Intention) UpdateIntention(workspaceId, intentionId db.Id, req *SaveIntentionRequest) (*IntentionResponse, error) {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}
	if req.Action != db.IntentionActionAllow && req.Action != db.IntentionActionDeny {
		return nil, ErrInvalidIntentionAction
	}

	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return nil, ErrIntentionNotFound
	}

	var skill *string
	if req.Destination.Skill != "" {
		value := req.Destination.Skill
		skill = &value
	}

	intention.SourceAgent = req.Source.Agent
	intention.DestinationAgent = req.Destination.Agent
	intention.Skill = skill
	intention.Action = req.Action
	err = i.IntentionRepository.Update(intention)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}

	intention, err = i.IntentionRepository.GetById(intention.Id)
	if err != nil || intention == nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}

	storedSkill := ""
	if intention.Skill != nil {
		storedSkill = *intention.Skill
	}

	return &IntentionResponse{
		Id:          intention.Id,
		WorkspaceId: intention.WorkspaceId,
		Source:      intention.SourceAgent,
		Destination: intention.DestinationAgent,
		Skill:       storedSkill,
		Action:      intention.Action,
		CreatedAt:   intention.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   intention.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// DeleteIntention removes a rule.
func (i *Intention) DeleteIntention(workspaceId, intentionId db.Id) error {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return ErrIntentionNotFound
	}

	err = i.IntentionRepository.Delete(intention.Id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntention, err)
	}

	return nil
}

// CheckIntention returns the matching rule, or deny when none exists.
func (i *Intention) CheckIntention(workspaceId db.Id, req *CheckIntentionRequest) (*CheckIntentionResponse, error) {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	intention, err := i.IntentionRepository.Match(
		workspaceId,
		req.Source.Agent,
		req.Destination.Agent,
		req.Destination.Skill,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCheckIntention, err)
	}
	if intention == nil {
		return &CheckIntentionResponse{Action: db.IntentionActionDeny}, nil
	}

	skill := ""
	if intention.Skill != nil {
		skill = *intention.Skill
	}

	return &CheckIntentionResponse{
		Action: intention.Action,
		Intention: &IntentionResponse{
			Id:          intention.Id,
			WorkspaceId: intention.WorkspaceId,
			Source:      intention.SourceAgent,
			Destination: intention.DestinationAgent,
			Skill:       skill,
			Action:      intention.Action,
			CreatedAt:   intention.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   intention.UpdatedAt.UTC().Format(time.RFC3339),
		},
	}, nil
}
