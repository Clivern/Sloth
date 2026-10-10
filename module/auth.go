// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"

	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountNotActive       = errors.New("account not active")
	ErrMaintenanceModeEnabled = errors.New("maintenance mode enabled")
	ErrFailedCheckMaintenance = errors.New("failed check maintenance")
	ErrFailedCreateSession    = errors.New("failed create session")

	ErrUserNotFound           = errors.New("user not found")
	ErrUserEmailAlreadyExists = errors.New("user with this email already exists")

	ErrFailedRevokeSessions = errors.New("failed revoke sessions")
)

// LoginResult is returned on a successful login.
type LoginResult struct {
	User          *db.User
	Session       *db.Session
	CookieOptions *util.CookieOptions
}

// Auth handles user authentication and sessions.
type Auth struct {
	UserRepository    db.UserRepository
	SessionRepository db.SessionRepository
	ConfigRepository  db.ConfigRepository
}

// NewAuth creates an Auth module with the given repositories.
func NewAuth(users db.UserRepository, sessions db.SessionRepository, config db.ConfigRepository) *Auth {
	return &Auth{
		UserRepository:    users,
		SessionRepository: sessions,
		ConfigRepository:  config,
	}
}

// Logout revokes all sessions for a user.
func (a *Auth) Logout(userId db.Id) error {
	count, err := a.SessionRepository.DeleteExpired()
	if err == nil && count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired sessions deleted")
	}

	err = a.SessionRepository.DeleteByUserId(userId)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedRevokeSessions, err)
	}

	log.Info().
		Str("userId", userId.String()).
		Msg("User logged out")

	return nil
}
