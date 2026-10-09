// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

// Package auth holds an agent's identity credential.
package auth

import (
	"crypto/subtle"
	"fmt"

	"github.com/clivern/cognit/pkg/util"
)

const (
	TypeAPIKey    = "api_key"
	TypeBasicAuth = "basic_auth"
)

// Auth is one identity credential: a type, a config map, and the key used
// to encrypt that config at rest.
type Auth struct {
	Type          string
	Config        map[string]string
	EncryptionKey string
}

// New returns an identity credential.
func New(authType string, config map[string]string, encryptionKey string) *Auth {
	return &Auth{
		Type:          authType,
		Config:        config,
		EncryptionKey: encryptionKey,
	}
}

// Encrypt seals each config value.
func (a *Auth) Encrypt() error {
	sealed := make(map[string]string, len(a.Config))
	for key, value := range a.Config {
		cipher, err := util.Encrypt(a.EncryptionKey, value)
		if err != nil {
			return fmt.Errorf("auth encrypt config: %w", err)
		}

		sealed[key] = cipher
	}

	a.Config = sealed

	return nil
}

// Validate decrypts the config, then checks type and required keys.
func (a *Auth) Validate() error {
	plain := make(map[string]string, len(a.Config))
	for key, value := range a.Config {
		decoded, err := util.Decrypt(a.EncryptionKey, value)
		if err != nil {
			return fmt.Errorf("auth decrypt config: %w", err)
		}

		plain[key] = decoded
	}

	a.Config = plain

	switch a.Type {
	case TypeAPIKey:
		if a.Config["value"] == "" {
			return fmt.Errorf("%w: %s needs a value", ErrConfig, TypeAPIKey)
		}
	case TypeBasicAuth:
		if a.Config["username"] == "" || a.Config["password"] == "" {
			return fmt.Errorf("%w: %s needs a username and a password", ErrConfig, TypeBasicAuth)
		}
	default:
		return fmt.Errorf("%w: %s", ErrType, a.Type)
	}

	return nil
}

// Match compares a presented credential to the decrypted config.
func (a *Auth) Match(presented map[string]string) error {
	switch a.Type {
	case TypeAPIKey:
		if subtle.ConstantTimeCompare([]byte(a.Config["value"]), []byte(presented["value"])) == 1 {
			return nil
		}
	case TypeBasicAuth:
		user := subtle.ConstantTimeCompare([]byte(a.Config["username"]), []byte(presented["username"]))
		pass := subtle.ConstantTimeCompare([]byte(a.Config["password"]), []byte(presented["password"]))
		if user == 1 && pass == 1 {
			return nil
		}
	}

	return ErrCredentials
}
