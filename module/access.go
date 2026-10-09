// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrAccessKeyNotFound           = errors.New("access key not found")
	ErrFailedCreateAccessKey       = errors.New("failed create access key")
	ErrFailedListAccessKeys        = errors.New("failed list access keys")
	ErrFailedGetAccessKey          = errors.New("failed get access key")
	ErrFailedDeleteAccessKey       = errors.New("failed delete access key")
	ErrInvalidAccessKeyPermissions = errors.New("invalid access key permissions")
)

// Access is the module for workspace access key CRUD.
type Access struct {
	AccessKeyRepository db.AccessKeyRepository
	WorkspaceRepository db.WorkspaceRepository
}

// NewAccess creates an access key module with the given repositories.
func NewAccess(accessKeys db.AccessKeyRepository, workspaces db.WorkspaceRepository) *Access {
	return &Access{
		AccessKeyRepository: accessKeys,
		WorkspaceRepository: workspaces,
	}
}

// ListAccessKeysResult is what you get back when listing workspace access keys.
type ListAccessKeysResult struct {
	Keys  []*AccessKeyResponse
	Total int64
}

// CreateAccessKeyRequest is what you pass when creating a workspace access key.
type CreateAccessKeyRequest struct {
	Name        string   `json:"name" validate:"required,max=60" label:"Name"`
	ExpiresAt   string   `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,required" label:"Permissions"`
}

// AccessKeyResponse is a workspace access key shaped for API responses.
type AccessKeyResponse struct {
	Id          db.Id    `json:"id"`
	WorkspaceId db.Id    `json:"workspaceId"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expiresAt,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Key         string   `json:"key,omitempty"`
}

// AccessKeyMeta is the JSON shape stored in access_keys.meta.
type AccessKeyMeta struct {
	Permissions []string `json:"permissions"`
}

// CreateAccessKey creates an access key; the raw key is returned once.
func (a *Access) CreateAccessKey(workspaceId db.Id, req *CreateAccessKeyRequest) (*AccessKeyResponse, error) {
	if !ValidAccessKeyPermissions(req.Permissions) {
		return nil, ErrInvalidAccessKeyPermissions
	}

	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	var expiresAt *time.Time
	if lo.IsNotEmpty(req.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, ErrInvalidExpiresAt
		}

		expiresAt = new(t)
	}

	raw, err := json.Marshal(AccessKeyMeta{Permissions: req.Permissions})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAccessKeyPermissions, err)
	}

	token, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateAccessKey, err)
	}

	item := &db.AccessKey{
		WorkspaceId: workspaceId,
		Name:        req.Name,
		Key:         token.String(),
		ExpiresAt:   expiresAt,
		Meta:        new(string(raw)),
	}

	err = a.AccessKeyRepository.Create(item)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateAccessKey, err)
	}

	resp := &AccessKeyResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Name:        item.Name,
		Permissions: req.Permissions,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
		Key:         token.String(),
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}

	return resp, nil
}

// ListAccessKeys returns workspace access keys (metadata only, never the secret).
func (a *Access) ListAccessKeys(workspaceId db.Id, limit, offset int) (*ListAccessKeysResult, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	if _, err := a.AccessKeyRepository.DeleteExpired(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired workspace access keys")
	}

	total, err := a.AccessKeyRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
	}

	keys, err := a.AccessKeyRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
	}

	list := make([]*AccessKeyResponse, 0, len(keys))
	for _, item := range keys {
		var meta AccessKeyMeta
		if err := json.Unmarshal([]byte(*item.Meta), &meta); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
		}

		resp := &AccessKeyResponse{
			Id:          item.Id,
			WorkspaceId: item.WorkspaceId,
			Name:        item.Name,
			Permissions: meta.Permissions,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: lo.TernaryF(
				item.ExpiresAt != nil,
				func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
		}
		list = append(list, resp)
	}

	return &ListAccessKeysResult{Keys: list, Total: total}, nil
}

// GetAccessKey returns one access key (never the secret).
func (a *Access) GetAccessKey(workspaceId, id db.Id) (*AccessKeyResponse, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	if _, err := a.AccessKeyRepository.DeleteExpired(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired workspace access keys")
	}

	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAccessKey, err)
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return nil, ErrAccessKeyNotFound
	}

	var meta AccessKeyMeta
	if err := json.Unmarshal([]byte(*item.Meta), &meta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAccessKey, err)
	}

	resp := &AccessKeyResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Name:        item.Name,
		Permissions: meta.Permissions,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}

	return resp, nil
}

// DeleteAccessKey deletes a workspace access key.
func (a *Access) DeleteAccessKey(workspaceId, id db.Id) error {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteAccessKey, err)
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return ErrAccessKeyNotFound
	}

	err = a.AccessKeyRepository.Delete(item.Id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedDeleteAccessKey, err)
	}

	return nil
}
