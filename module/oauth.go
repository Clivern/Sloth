// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/cognit/conf"
	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// OAuthIdentity is a verified identity from an OAuth provider.
type OAuthIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
}

// LoginWithOAuth finds or creates a user from an OAuth identity and opens a session.
func (a *Auth) LoginWithOAuth(ctx context.Context, identity *OAuthIdentity) (*LoginResult, error) {
	user, err := a.UserRepository.GetByProvider(identity.Provider, identity.ProviderUserID)
	if err != nil {
		return nil, fmt.Errorf("get user by provider: %w", err)
	}

	foundByProvider := user != nil

	if lo.IsEmpty(identity.Email) {
		return nil, fmt.Errorf("oauth identity missing email")
	}

	if user == nil {
		user, err = a.UserRepository.GetByEmail(identity.Email)
		if err != nil {
			return nil, fmt.Errorf("get user by email: %w", err)
		}
	}

	if user != nil && !foundByProvider && (user.Provider == db.UserProviderLocal || lo.IsEmpty(user.Provider)) {
		providerUserID := identity.ProviderUserID
		user.Provider = identity.Provider
		user.ProviderUserId = &providerUserID
		user.IsEmailVerified = true
		if lo.IsNotEmpty(identity.Name) {
			user.Name = identity.Name
		}

		err = a.UserRepository.Update(user)
		if err != nil {
			return nil, fmt.Errorf("convert local user to oauth: %w", err)
		}
	}

	if user == nil {
		pwd, err := util.GenerateSecureToken(20)
		if err != nil {
			return nil, fmt.Errorf("generate oauth password: %w", err)
		}

		hpass, err := util.HashPassword(pwd)
		if err != nil {
			return nil, fmt.Errorf("hash oauth password: %w", err)
		}

		now := time.Now().UTC()
		providerUserID := identity.ProviderUserID
		user = &db.User{
			Name:            identity.Name,
			Email:           identity.Email,
			Password:        hpass,
			Provider:        identity.Provider,
			ProviderUserId:  &providerUserID,
			Role:            db.UserRoleRegular,
			IsActive:        true,
			IsEmailVerified: true,
			LastLoginAt:     now,
			Language:        db.UserLanguageEN,
			Theme:           db.UserThemeDefault,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		err = a.UserRepository.Create(user)
		if err != nil {
			return nil, fmt.Errorf("create oauth user: %w", err)
		}
	}

	if foundByProvider && lo.IsNotEmpty(identity.Email) && identity.Email != user.Email {
		existing, err := a.UserRepository.GetByEmail(identity.Email)
		if err != nil {
			return nil, fmt.Errorf("sync oauth email: %w", err)
		}
		if existing != nil && existing.Id != user.Id {
			return nil, ErrUserEmailAlreadyExists
		}

		user.Email = identity.Email
		user.IsEmailVerified = true
		err = a.UserRepository.Update(user)
		if err != nil {
			return nil, fmt.Errorf("sync oauth email: %w", err)
		}
	}

	if db.IsBotUser(user.Id) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountNotActive
	}

	maintenance, err := a.ConfigRepository.Get("maintenance_mode")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCheckMaintenance, err)
	}
	if maintenance != nil && maintenance.Value == "1" && user.Role != db.UserRoleAdmin {
		return nil, ErrMaintenanceModeEnabled
	}

	err = a.UserRepository.UpdateLastLogin(user.Id)
	if err != nil {
		return nil, fmt.Errorf("update last login: %w", err)
	}

	token, err := util.GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateSession, err)
	}

	session := &db.Session{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: time.Now().UTC().Add(conf.DefaultSessionDuration),
	}

	if appCtx, ok := AppCtxFrom(ctx); ok {
		if lo.IsNotEmpty(appCtx.IPAddress) {
			session.IPAddress = &appCtx.IPAddress
		}
		if lo.IsNotEmpty(appCtx.UserAgent) {
			session.UserAgent = &appCtx.UserAgent
		}
	}

	err = a.SessionRepository.Create(session)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateSession, err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("sessionId", session.Id.String()).
		Str("provider", identity.Provider).
		Msg("User logged in via oauth")

	cookieOptions := lo.Ternary(
		strings.HasPrefix(viper.GetString("app.url"), "https://"),
		util.SecureCookieOptions(),
		util.DefaultCookieOptions(),
	)

	return &LoginResult{
		User:          user,
		Session:       session,
		CookieOptions: cookieOptions,
	}, nil
}
