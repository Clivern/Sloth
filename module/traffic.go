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
	ErrTrafficNotFound   = errors.New("traffic not found")
	ErrFailedListTraffic = errors.New("failed list traffic")
	ErrFailedGetTraffic  = errors.New("failed get traffic")
)

// Traffic is the module for recorded gateway calls.
type Traffic struct {
	TrafficRepository   db.TrafficRepository
	WorkspaceRepository db.WorkspaceRepository
}

// TrafficCallResponse is one gateway call shaped for API responses.
type TrafficCallResponse struct {
	Id                  db.Id  `json:"id"`
	WorkspaceId         db.Id  `json:"workspaceId"`
	SourceInstance      string `json:"sourceInstance"`
	DestinationInstance string `json:"destinationInstance"`
	Skill               string `json:"skill"`
	Verb                string `json:"verb"`
	Decision            string `json:"decision"`
	Result              string `json:"result"`
	LatencyMs           int64  `json:"latencyMs"`
	TaskId              string `json:"taskId,omitempty"`
	CreatedAt           string `json:"createdAt"`
}

// ListTrafficResponse is returned when listing gateway calls.
type ListTrafficResponse struct {
	Calls []*TrafficCallResponse
	Total int64
}

// NewTraffic creates a traffic module with the given repositories.
func NewTraffic(calls db.TrafficRepository, workspaces db.WorkspaceRepository) *Traffic {
	return &Traffic{
		TrafficRepository:   calls,
		WorkspaceRepository: workspaces,
	}
}

// ListTraffic returns recorded gateway calls for a workspace.
func (t *Traffic) ListTraffic(workspaceId db.Id, limit, offset int) (*ListTrafficResponse, error) {
	workspace, err := t.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	total, err := t.TrafficRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListTraffic, err)
	}

	calls, err := t.TrafficRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListTraffic, err)
	}

	list := make([]*TrafficCallResponse, 0, len(calls))
	for _, call := range calls {
		taskId := ""
		if call.TaskId != nil {
			taskId = *call.TaskId
		}

		list = append(list, &TrafficCallResponse{
			Id:                  call.Id,
			WorkspaceId:         call.WorkspaceId,
			SourceInstance:      call.SourceInstance,
			DestinationInstance: call.DestinationInstance,
			Skill:               call.Skill,
			Verb:                call.Verb,
			Decision:            call.Decision,
			Result:              call.Result,
			LatencyMs:           int64(call.LatencyMs),
			TaskId:              taskId,
			CreatedAt:           call.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListTrafficResponse{Calls: list, Total: total}, nil
}

// GetTraffic returns one gateway call by id.
func (t *Traffic) GetTraffic(workspaceId, trafficId db.Id) (*TrafficCallResponse, error) {
	workspace, err := t.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	call, err := t.TrafficRepository.GetById(trafficId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetTraffic, err)
	}
	if call == nil || call.WorkspaceId != workspaceId {
		return nil, ErrTrafficNotFound
	}

	taskId := ""
	if call.TaskId != nil {
		taskId = *call.TaskId
	}

	return &TrafficCallResponse{
		Id:                  call.Id,
		WorkspaceId:         call.WorkspaceId,
		SourceInstance:      call.SourceInstance,
		DestinationInstance: call.DestinationInstance,
		Skill:               call.Skill,
		Verb:                call.Verb,
		Decision:            call.Decision,
		Result:              call.Result,
		LatencyMs:           int64(call.LatencyMs),
		TaskId:              taskId,
		CreatedAt:           call.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}
