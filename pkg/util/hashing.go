// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash from a plain text password.
// It uses the default cost factor (bcrypt.DefaultCost = 10).
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a bcrypt hashed password with a plain text password.
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(password),
	)

	return err == nil
}

// RandomHash returns a 12-character cryptographically secure random hex string.
func RandomHash() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// MapChecksum returns a deterministic SHA-256 hex checksum of a map.
func MapChecksum(m map[string]any) (string, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	hasher := sha256.New()
	for _, k := range keys {
		valBytes, err := json.Marshal(m[k])
		if err != nil {
			return "", fmt.Errorf("failed to marshal value for key %s: %w", k, err)
		}

		hasher.Write([]byte(k))
		hasher.Write(valBytes)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
