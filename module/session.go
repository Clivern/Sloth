// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// SessionManager manages sessions (create, revoke, etc.).
type SessionManager struct {
	SessionRepo db.SessionRepository
	UserRepo    db.UserRepository
}

// NewSessionManager creates a new session manager.
func NewSessionManager(sessions db.SessionRepository, users db.UserRepository) *SessionManager {
	return &SessionManager{
		SessionRepo: sessions,
		UserRepo:    users,
	}
}

// CreateSession creates a new session for a user. TODO: remove if not used in future.
func (s *SessionManager) CreateSession(ctx context.Context, userId db.Id, duration time.Duration) (*db.Session, error) {
	user, err := s.UserRepo.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	token, err := util.GenerateSecureToken(32)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to generate session token")
		return nil, err
	}

	session := &db.Session{
		Token:     token,
		UserId:    userId,
		ExpiresAt: time.Now().UTC().Add(duration),
	}
	if appCtx, ok := AppCtxFrom(ctx); ok && appCtx != nil {
		if lo.IsNotEmpty(appCtx.IPAddress) {
			session.IPAddress = &appCtx.IPAddress
		}
		if lo.IsNotEmpty(appCtx.UserAgent) {
			session.UserAgent = &appCtx.UserAgent
		}
	}

	err = s.SessionRepo.Create(session)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to create session")
		return nil, err
	}

	log.Info().
		Str("userId", userId.String()).
		Str("sessionId", session.Id.String()).
		Msg("Session created")

	return session, nil
}

// ValidateSession validates a session token and returns the associated user.
func (s *SessionManager) ValidateSession(token string) (*db.User, *db.Session, error) {
	session, err := s.SessionRepo.GetByToken(token)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get session")
		return nil, nil, err
	}
	if session == nil {
		return nil, nil, errors.New("session not found")
	}
	if session.ExpiresAt.Before(time.Now().UTC()) {
		err = s.SessionRepo.Delete(session.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("sessionId", session.Id.String()).
				Msg("Failed to delete expired session")
		}
		log.Debug().
			Str("sessionId", session.Id.String()).
			Str("userId", session.UserId.String()).
			Msg("Session expired")
		return nil, nil, errors.New("session expired")
	}

	user, err := s.UserRepo.GetById(session.UserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionId", session.Id.String()).
			Str("userId", session.UserId.String()).
			Msg("Failed to get session user")
		return nil, nil, err
	}
	if user == nil {
		log.Warn().
			Str("sessionId", session.Id.String()).
			Str("userId", session.UserId.String()).
			Msg("Session user not found")
		return nil, nil, errors.New("user not found")
	}
	if !user.IsActive {
		log.Debug().
			Str("sessionId", session.Id.String()).
			Str("userId", user.Id.String()).
			Msg("Session rejected for inactive user")
		return nil, nil, errors.New("user is not active")
	}

	return user, session, nil
}

// RevokeUserSessions revokes all sessions for a user.
func (s *SessionManager) RevokeUserSessions(userId db.Id) error {
	err := s.SessionRepo.DeleteByUserId(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to revoke user sessions")
		return err
	}

	log.Info().
		Str("userId", userId.String()).
		Msg("User sessions revoked")

	return nil
}

// GetUserSessions retrieves all active sessions for a user. TODO: remove if not used in future.
func (s *SessionManager) GetUserSessions(userId db.Id) ([]*db.Session, error) {
	sessions, err := s.SessionRepo.GetByUserId(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to list user sessions")
		return nil, err
	}

	return lo.Filter(sessions, func(session *db.Session, _ int) bool {
		return session.ExpiresAt.After(time.Now().UTC())
	}), nil
}

// CleanupExpiredSessions removes all expired sessions from the database. TODO: remove if not used in future.
func (s *SessionManager) CleanupExpiredSessions() (int64, error) {
	count, err := s.SessionRepo.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired sessions")
		return 0, err
	}

	if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired sessions deleted")
	}

	return count, nil
}
