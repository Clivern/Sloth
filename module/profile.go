// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrFailedGetProfile    = errors.New("failed get profile")
	ErrFailedUpdateProfile = errors.New("failed update profile")
)

// UpdateProfileRequest is the body for updating your profile.
type UpdateProfileRequest struct {
	Name     string `json:"name" validate:"omitempty,max=255" label:"Name"`
	Language string `json:"language" validate:"required,oneof=en fr" label:"Language"`
	Theme    string `json:"theme" validate:"required,oneof=default blue slate emerald dark" label:"Theme"`
}

// Profile handles profile updates.
type Profile struct {
	UserRepository db.UserRepository
}

// NewProfile creates a Profile module with the given repositories.
func NewProfile(users db.UserRepository) *Profile {
	return &Profile{
		UserRepository: users,
	}
}

// GetProfile retrieves a user's profile.
func (p *Profile) GetProfile(userId db.Id) (*db.User, error) {
	user, err := p.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return nil, fmt.Errorf("%w: get user: %v", ErrFailedGetProfile, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateProfile updates a user's profile.
func (p *Profile) UpdateProfile(userId db.Id, req *UpdateProfileRequest) (*db.User, error) {
	user, err := p.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return nil, fmt.Errorf("%w: get user: %v", ErrFailedUpdateProfile, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.Name = lo.Ternary(lo.IsNotEmpty(req.Name), req.Name, user.Name)
	user.Language = req.Language
	user.Theme = req.Theme

	err = p.UserRepository.Update(user)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to update profile")
		return nil, fmt.Errorf("%w: update user: %v", ErrFailedUpdateProfile, err)
	}

	log.Info().
		Str("userId", userId.String()).
		Str("language", user.Language).
		Str("theme", user.Theme).
		Msg("Profile updated")

	return user, nil
}
