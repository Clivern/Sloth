// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrInvalidExpiresAt = errors.New("invalid expires at format")
)

// APIKey is the module for user API key CRUD.
type APIKey struct {
	APIKeyRepository db.APIKeyRepository
}

// NewAPIKey creates an API key module with the given repository.
func NewAPIKey(apiKeys db.APIKeyRepository) *APIKey {
	return &APIKey{APIKeyRepository: apiKeys}
}

// ListAPIKeysResult is what you get back when listing API keys.
type ListAPIKeysResult struct {
	APIKeys []*APIKeyResponse
	Total   int64
}

// CreateAPIKeyRequest is what you pass when creating an API key.
type CreateAPIKeyRequest struct {
	Name      string `json:"name" validate:"required,max=255" label:"Name"`
	ExpiresAt string `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
}

// APIKeyResponse is an API key shaped for API responses.
type APIKeyResponse struct {
	Id        db.Id   `json:"id"`
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expiresAt"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	Key       string  `json:"key,omitempty"`
}

// CreateAPIKey creates an API key for the user; the raw key is returned once.
func (a *APIKey) CreateAPIKey(req *CreateAPIKeyRequest, user *db.User) (*APIKeyResponse, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	var expiresAt *time.Time
	if lo.IsNotEmpty(req.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, ErrInvalidExpiresAt
		}

		expiresAt = new(t)
	}

	apiKey := &db.APIKey{
		UserId:    user.Id,
		Name:      req.Name,
		Key:       id.String(),
		ExpiresAt: expiresAt,
	}

	err = a.APIKeyRepository.Create(apiKey)
	if err != nil {
		return nil, err
	}

	return &APIKeyResponse{
		Id:        apiKey.Id,
		Name:      req.Name,
		CreatedAt: apiKey.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: apiKey.UpdatedAt.UTC().Format(time.RFC3339),
		Key:       id.String(),
		ExpiresAt: lo.TernaryF(
			apiKey.ExpiresAt != nil,
			func() *string { return new(apiKey.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}, nil
}

// ListAPIKeys returns the user's API keys (metadata only, never the secret).
func (a *APIKey) ListAPIKeys(user *db.User, limit, offset int) (*ListAPIKeysResult, error) {
	count, err := a.APIKeyRepository.DeleteExpired()
	if err == nil && count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired API keys deleted")
	}

	total, err := a.APIKeyRepository.CountByUserId(user.Id)
	if err != nil {
		return nil, err
	}

	keys, err := a.APIKeyRepository.ListByUserId(user.Id, limit, offset)
	if err != nil {
		return nil, err
	}

	list := make([]*APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		list = append(list, &APIKeyResponse{
			Id:        k.Id,
			Name:      k.Name,
			CreatedAt: k.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: k.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: lo.TernaryF(
				k.ExpiresAt != nil,
				func() *string { return new(k.ExpiresAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
		})
	}

	return &ListAPIKeysResult{APIKeys: list, Total: total}, nil
}

// GetAPIKey returns a single API key owned by the user (never the secret).
func (a *APIKey) GetAPIKey(id db.Id, user *db.User) (*APIKeyResponse, error) {
	count, err := a.APIKeyRepository.DeleteExpired()
	if err == nil && count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired API keys deleted")
	}

	k, err := a.APIKeyRepository.GetById(id)
	if err != nil {
		return nil, err
	}
	if k == nil || k.UserId != user.Id {
		return nil, ErrAPIKeyNotFound
	}

	return &APIKeyResponse{
		Id:        k.Id,
		Name:      k.Name,
		CreatedAt: k.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: k.UpdatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: lo.TernaryF(
			k.ExpiresAt != nil,
			func() *string { return new(k.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}, nil
}

// DeleteAPIKey deletes a single API key owned by the user.
func (a *APIKey) DeleteAPIKey(id db.Id, user *db.User) error {
	k, err := a.APIKeyRepository.GetById(id)
	if err != nil {
		return err
	}
	if k == nil || k.UserId != user.Id {
		return ErrAPIKeyNotFound
	}

	return a.APIKeyRepository.Delete(k.Id)
}
