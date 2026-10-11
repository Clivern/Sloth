// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/encrypt"

	"github.com/samber/lo"
)

var (
	ErrProtectionNotFound      = errors.New("protection not found")
	ErrInvalidProtectionConfig = errors.New("invalid protection config")
)

// ProtectionConfig is the JSON shape stored in agent_protection.config.
type ProtectionConfig struct {
	Header          string `json:"header,omitempty"`
	Username        string `json:"username,omitempty"`
	SecretEncrypted string `json:"secretEncrypted,omitempty"`
}

// SaveProtectionRequest is the body for setting or replacing an agent protection.
type SaveProtectionRequest struct {
	Type     string `json:"type" validate:"required,oneof=api_key basic_auth" label:"Type"`
	Username string `json:"username" validate:"omitempty,max=100" label:"Username"`
	Header   string `json:"header" validate:"omitempty,max=100" label:"Header"`
	Secret   string `json:"secret" validate:"omitempty,min=8,max=512" label:"Secret"`
}

// UpdateProtectionStatusRequest is the body for turning an agent protection on or off.
type UpdateProtectionStatusRequest struct {
	IsActive bool `json:"isActive" label:"Active"`
}

// ProtectionResponse is an agent protection shaped for API responses.
type ProtectionResponse struct {
	Id        db.Id  `json:"id"`
	Type      string `json:"type"`
	Username  string `json:"username,omitempty"`
	Header    string `json:"header,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Secret    string `json:"secret,omitempty"`
}

// Protection is the module for the credential the gateway uses to reach an agent.
type Protection struct {
	AgentRepository      db.AgentRepository
	ProtectionRepository db.AgentProtectionRepository
}

// NewProtection creates a protection module with the given repositories.
func NewProtection(agents db.AgentRepository, protections db.AgentProtectionRepository) *Protection {
	return &Protection{
		AgentRepository:      agents,
		ProtectionRepository: protections,
	}
}

// GetProtection returns the protection of an agent, never its secret.
func (p *Protection) GetProtection(workspaceId db.Id, agentName string) (*ProtectionResponse, error) {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}
	if protection == nil {
		return nil, ErrProtectionNotFound
	}

	var config ProtectionConfig
	_ = json.Unmarshal([]byte(protection.Config), &config)

	return &ProtectionResponse{
		Id:        protection.Id,
		Type:      protection.Type,
		Username:  config.Username,
		Header:    config.Header,
		IsActive:  protection.IsActive,
		CreatedAt: protection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: protection.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// SaveProtection sets or replaces the protection of an agent; a generated secret is returned once.
func (p *Protection) SaveProtection(workspaceId db.Id, agentName string, req *SaveProtectionRequest) (*ProtectionResponse, error) {
	if (req.Type == db.AgentProtectionTypeBasicAuth) == (req.Username == "") {
		return nil, ErrInvalidProtectionConfig
	}

	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	secret := lo.Ternary(req.Secret == "", rand.Text(), req.Secret)
	config := ProtectionConfig{Username: req.Username}

	switch req.Type {
	case db.AgentProtectionTypeAPIKey:
		config.Header = lo.Ternary(req.Header == "", "X-API-Key", req.Header)
		config.SecretEncrypted, err = encrypt.Encrypt(secret)
	case db.AgentProtectionTypeBasicAuth:
		config.SecretEncrypted, err = encrypt.Encrypt(secret)
	}
	if err != nil {
		return nil, err
	}

	raw, _ := json.Marshal(config)

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}

	if protection == nil {
		protection = &db.AgentProtection{AgentId: agent.Id, Type: req.Type, Config: string(raw), IsActive: true}
		err = p.ProtectionRepository.Create(protection)
	} else {
		protection.Type = req.Type
		protection.Config = string(raw)
		protection.IsActive = true
		err = p.ProtectionRepository.Update(protection)
	}
	if err != nil {
		return nil, err
	}

	return &ProtectionResponse{
		Id:        protection.Id,
		Type:      protection.Type,
		Username:  config.Username,
		Header:    config.Header,
		IsActive:  protection.IsActive,
		CreatedAt: protection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: protection.UpdatedAt.UTC().Format(time.RFC3339),
		Secret:    lo.Ternary(req.Secret == "", secret, ""),
	}, nil
}

// UpdateProtectionStatus turns the protection of an agent on or off.
func (p *Protection) UpdateProtectionStatus(workspaceId db.Id, agentName string, req *UpdateProtectionStatusRequest) (*ProtectionResponse, error) {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}
	if protection == nil {
		return nil, ErrProtectionNotFound
	}

	protection.IsActive = req.IsActive

	err = p.ProtectionRepository.Update(protection)
	if err != nil {
		return nil, err
	}

	var config ProtectionConfig
	_ = json.Unmarshal([]byte(protection.Config), &config)

	return &ProtectionResponse{
		Id:        protection.Id,
		Type:      protection.Type,
		Username:  config.Username,
		Header:    config.Header,
		IsActive:  protection.IsActive,
		CreatedAt: protection.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: protection.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// DeleteProtection removes the protection of an agent.
func (p *Protection) DeleteProtection(workspaceId db.Id, agentName string) error {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return err
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		return err
	}
	if protection == nil {
		return ErrProtectionNotFound
	}

	return p.ProtectionRepository.Delete(protection.Id)
}
