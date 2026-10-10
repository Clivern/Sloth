// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/samber/lo"
)

// Me resolves session, API key and access key principals for /me.
type Me struct {
	APIKeyRepository    db.APIKeyRepository
	UserRepository      db.UserRepository
	AccessKeyRepository db.AccessKeyRepository
}

// MeAPIKeyOwner is the user who owns an API key.
type MeAPIKeyOwner struct {
	Id    db.Id  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// MeAPIKeyInfo is API key metadata returned by /me.
type MeAPIKeyInfo struct {
	Id        db.Id   `json:"id"`
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expiresAt,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// MeAPIKeyResponse is returned when /me is called with X-API-Key.
type MeAPIKeyResponse struct {
	PrincipalType string        `json:"principalType"`
	Owner         MeAPIKeyOwner `json:"owner"`
	APIKey        MeAPIKeyInfo  `json:"apiKey"`
}

// MeAccessKeyResponse is returned when /me is called with X-Access-Key.
type MeAccessKeyResponse struct {
	PrincipalType string            `json:"principalType"`
	AccessKey     AccessKeyResponse `json:"accessKey"`
}

// MeSessionResponse is returned when /me is called with a session cookie.
type MeSessionResponse struct {
	PrincipalType string        `json:"principalType"`
	User          MeAPIKeyOwner `json:"user"`
}

// NewMe creates a /me module with the given repositories.
func NewMe(apiKeys db.APIKeyRepository, users db.UserRepository, accessKeys db.AccessKeyRepository) *Me {
	return &Me{
		APIKeyRepository:    apiKeys,
		UserRepository:      users,
		AccessKeyRepository: accessKeys,
	}
}

// GetBySession returns the session user.
func (m *Me) GetBySession(user *db.User) *MeSessionResponse {
	return &MeSessionResponse{
		PrincipalType: "session",
		User: MeAPIKeyOwner{
			Id:    user.Id,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}
}

// GetByAPIKey returns the API key owner and key metadata.
func (m *Me) GetByAPIKey(key string) (*MeAPIKeyResponse, error) {
	apiKey, err := m.APIKeyRepository.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}

	owner, err := m.UserRepository.GetById(apiKey.UserId)
	if err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, ErrAPIKeyNotFound
	}

	return &MeAPIKeyResponse{
		PrincipalType: "api_key",
		Owner: MeAPIKeyOwner{
			Id:    owner.Id,
			Email: owner.Email,
			Name:  owner.Name,
			Role:  owner.Role,
		},
		APIKey: MeAPIKeyInfo{
			Id:        apiKey.Id,
			Name:      apiKey.Name,
			CreatedAt: apiKey.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: apiKey.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: lo.TernaryF(
				apiKey.ExpiresAt != nil,
				func() *string { return new(apiKey.ExpiresAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
		},
	}, nil
}

// GetByAccessKey returns workspace access key metadata and permissions.
func (m *Me) GetByAccessKey(key string) (*MeAccessKeyResponse, error) {
	accessKey, err := m.AccessKeyRepository.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if accessKey == nil {
		return nil, ErrAccessKeyNotFound
	}

	var meta AccessKeyMeta
	err = json.Unmarshal([]byte(*accessKey.Meta), &meta)
	if err != nil {
		return nil, err
	}

	return &MeAccessKeyResponse{
		PrincipalType: "access_key",
		AccessKey: AccessKeyResponse{
			Id:          accessKey.Id,
			WorkspaceId: accessKey.WorkspaceId,
			Name:        accessKey.Name,
			Permissions: meta.Permissions,
			CreatedAt:   accessKey.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   accessKey.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: lo.TernaryF(
				accessKey.ExpiresAt != nil,
				func() *string { return new(accessKey.ExpiresAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
		},
	}, nil
}
