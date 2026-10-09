// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
)

var (
	ErrPlatformAlreadyInstalled = errors.New("platform already installed")
	ErrFailedCompleteSetup      = errors.New("failed complete setup")
)

// Setup does the initial platform install and config.
type Setup struct {
	ConfigRepository db.ConfigRepository
	UserRepository   db.UserRepository
}

// SetupRequest is the body you send when running setup.
type SetupRequest struct {
	PlatformEmail string `json:"platformEmail" validate:"required,email,min=4,max=60" label:"Platform Email"`
}

// NewSetup creates a Setup with the given repositories.
func NewSetup(config db.ConfigRepository, users db.UserRepository) *Setup {
	return &Setup{
		ConfigRepository: config,
		UserRepository:   users,
	}
}

// IsInstalled checks whether the platform has been installed.
func (s *Setup) IsInstalled() bool {
	config, err := s.ConfigRepository.Get("is_installed")
	if err != nil {
		return false
	}

	return config != nil
}

// Install performs the initial platform installation with the provided request.
func (s *Setup) Install(req *SetupRequest) error {
	if s.IsInstalled() {
		return ErrPlatformAlreadyInstalled
	}

	botPassword, err := util.HashPassword("cognit-bot-user-no-login")
	if err != nil {
		return fmt.Errorf("%w: hash bot password: %v", ErrFailedCompleteSetup, err)
	}

	err = s.UserRepository.Create(&db.User{
		Id:              db.BotUserId,
		Name:            db.BotUserName,
		Email:           req.PlatformEmail,
		Password:        botPassword,
		Provider:        db.UserProviderLocal,
		Role:            db.UserRoleBot,
		IsActive:        true,
		IsEmailVerified: true,
		LastLoginAt:     time.Now().UTC(),
		Language:        db.UserLanguageEN,
		Theme:           db.UserThemeDefault,
	})
	if err != nil {
		return fmt.Errorf("%w: create bot user: %v", ErrFailedCompleteSetup, err)
	}

	err = s.ConfigRepository.Create("is_installed", "1")
	if err != nil {
		return fmt.Errorf("%w: create is_installed config: %v", ErrFailedCompleteSetup, err)
	}

	err = s.ConfigRepository.Create("platform_email", req.PlatformEmail)
	if err != nil {
		return fmt.Errorf("%w: create platform_email config: %v", ErrFailedCompleteSetup, err)
	}

	err = s.ConfigRepository.Create(
		"platform_description",
		"A Service Discovery Registry for AI Agents",
	)
	if err != nil {
		return fmt.Errorf("%w: create platform_description config: %v", ErrFailedCompleteSetup, err)
	}

	err = s.ConfigRepository.Create("maintenance_mode", "0")
	if err != nil {
		return fmt.Errorf("%w: create maintenance_mode config: %v", ErrFailedCompleteSetup, err)
	}

	return nil
}
