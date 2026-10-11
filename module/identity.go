// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
)

var (
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrIdentityAlreadyExists = errors.New("identity already exists")
	ErrInvalidIdentityConfig = errors.New("invalid identity config")
)

// IdentityConfig is the JSON shape stored in agent_identity.config.
type IdentityConfig struct {
	Username   string `json:"username,omitempty"`
	SecretHash string `json:"secretHash"`
}

// CreateIdentityRequest is the body for creating an agent identity.
type CreateIdentityRequest struct {
	Name     string `json:"name" validate:"required,max=100" label:"Name"`
	Type     string `json:"type" validate:"required,oneof=api_key basic_auth" label:"Type"`
	Username string `json:"username" validate:"omitempty,max=100" label:"Username"`
}

// UpdateIdentityRequest is the body for updating an agent identity.
type UpdateIdentityRequest struct {
	Name     string `json:"name" validate:"required,max=100" label:"Name"`
	IsActive bool   `json:"isActive" label:"Active"`
}

// IdentityResponse is an agent identity shaped for API responses.
type IdentityResponse struct {
	Id        db.Id  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Username  string `json:"username,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Secret    string `json:"secret,omitempty"`
}

// Identity is the module for the credentials an agent uses to call through the gateway.
type Identity struct {
	AgentRepository    db.AgentRepository
	IdentityRepository db.AgentIdentityRepository
}

// NewIdentity creates an identity module with the given repositories.
func NewIdentity(agents db.AgentRepository, identities db.AgentIdentityRepository) *Identity {
	return &Identity{
		AgentRepository:    agents,
		IdentityRepository: identities,
	}
}

// ListIdentities returns the identities of an agent, never their secrets.
func (i *Identity) ListIdentities(workspaceId db.Id, agentName string) ([]*IdentityResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	identities, err := i.IdentityRepository.ListByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}

	list := make([]*IdentityResponse, 0, len(identities))
	for _, identity := range identities {
		var config IdentityConfig
		_ = json.Unmarshal([]byte(identity.Config), &config)

		list = append(list, &IdentityResponse{
			Id:        identity.Id,
			Name:      identity.Name,
			Type:      identity.Type,
			Username:  config.Username,
			IsActive:  identity.IsActive,
			CreatedAt: identity.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: identity.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return list, nil
}

// CreateIdentity issues a new identity for an agent; the raw secret is returned once.
func (i *Identity) CreateIdentity(
	workspaceId db.Id,
	agentName string,
	req *CreateIdentityRequest,
) (*IdentityResponse, error) {
	if (req.Type == db.AgentIdentityTypeBasicAuth) == (req.Username == "") {
		return nil, ErrInvalidIdentityConfig
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	existing, err := i.IdentityRepository.GetByAgentAndName(agent.Id, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrIdentityAlreadyExists
	}

	secret := rand.Text()

	hash, err := util.HashPassword(secret)
	if err != nil {
		return nil, err
	}

	raw, _ := json.Marshal(IdentityConfig{Username: req.Username, SecretHash: hash})

	identity := &db.AgentIdentity{
		AgentId:  agent.Id,
		Name:     req.Name,
		Type:     req.Type,
		Config:   string(raw),
		IsActive: true,
	}

	err = i.IdentityRepository.Create(identity)
	if err != nil {
		return nil, err
	}

	return &IdentityResponse{
		Id:        identity.Id,
		Name:      identity.Name,
		Type:      identity.Type,
		Username:  req.Username,
		IsActive:  identity.IsActive,
		CreatedAt: identity.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: identity.UpdatedAt.UTC().Format(time.RFC3339),
		Secret:    secret,
	}, nil
}

// UpdateIdentity renames an agent identity or turns it on or off.
func (i *Identity) UpdateIdentity(
	workspaceId db.Id,
	agentName string,
	id db.Id,
	req *UpdateIdentityRequest,
) (*IdentityResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	identity, err := i.IdentityRepository.GetById(id)
	if err != nil {
		return nil, err
	}
	if identity == nil || identity.AgentId != agent.Id {
		return nil, ErrIdentityNotFound
	}

	if req.Name != identity.Name {
		existing, err := i.IdentityRepository.GetByAgentAndName(agent.Id, req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrIdentityAlreadyExists
		}
	}

	identity.Name = req.Name
	identity.IsActive = req.IsActive

	err = i.IdentityRepository.Update(identity)
	if err != nil {
		return nil, err
	}

	var config IdentityConfig
	_ = json.Unmarshal([]byte(identity.Config), &config)

	return &IdentityResponse{
		Id:        identity.Id,
		Name:      identity.Name,
		Type:      identity.Type,
		Username:  config.Username,
		IsActive:  identity.IsActive,
		CreatedAt: identity.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// DeleteIdentity revokes an agent identity.
func (i *Identity) DeleteIdentity(workspaceId db.Id, agentName string, id db.Id) error {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return err
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	identity, err := i.IdentityRepository.GetById(id)
	if err != nil {
		return err
	}
	if identity == nil || identity.AgentId != agent.Id {
		return ErrIdentityNotFound
	}

	return i.IdentityRepository.Delete(identity.Id)
}
