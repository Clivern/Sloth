// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUnitPasswordHashing(t *testing.T) {
	t.Run("HashPassword", func(t *testing.T) {
		passwords := []string{
			"mySecurePassword123",
			"",
			"thisIsAVeryLongPasswordWithCharacters1234567890!@#$%^&*()_+-=[]",
			"p@ssw0rd!#$%^&*()",
		}

		for _, password := range passwords {
			t.Run(password, func(t *testing.T) {
				hashed, err := HashPassword(password)
				assert.NoError(t, err)
				assert.NotEmpty(t, hashed)
				assert.NotEqual(t, password, hashed)
			})
		}

		hash1, err1 := HashPassword("samePassword")
		hash2, err2 := HashPassword("samePassword")
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("ComparePassword", func(t *testing.T) {
		cases := []struct {
			name     string
			password string
			compare  string
			hash     string
			expected bool
		}{
			{
				name:     "matching password",
				password: "mySecurePassword123",
				compare:  "mySecurePassword123",
				expected: true,
			},
			{
				name:     "non-matching password",
				password: "correctPassword",
				compare:  "wrongPassword",
				expected: false,
			},
			{
				name:     "empty password",
				password: "somePassword",
				compare:  "",
				expected: false,
			},
			{
				name:     "case sensitivity",
				password: "Password123",
				compare:  "password123",
				expected: false,
			},
			{
				name:     "special characters",
				password: "p@ssw0rd!#$%",
				compare:  "p@ssw0rd!#$%",
				expected: true,
			},
			{
				name:     "invalid hash",
				hash:     "notAValidBcryptHash",
				compare:  "somePassword",
				expected: false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				hash := tc.hash
				if lo.IsEmpty(hash) {
					var err error
					hash, err = HashPassword(tc.password)
					assert.NoError(t, err)
				}

				assert.Equal(t, tc.expected, ComparePassword(hash, tc.compare))
			})
		}
	})

	t.Run("Complete password workflow", func(t *testing.T) {
		passwords := []string{
			"simplePassword",
			"Complex123!@#",
			"",
			"with spaces in password",
			"unicode密码🔐",
		}

		for _, password := range passwords {
			t.Run("Password: "+password, func(t *testing.T) {
				hashed, err := HashPassword(password)
				assert.NoError(t, err)
				assert.True(t, ComparePassword(hashed, password))
				assert.False(t, ComparePassword(hashed, password+"wrong"))
			})
		}
	})

	t.Run("RandomHash", func(t *testing.T) {
		hash, err := RandomHash()
		assert.NoError(t, err)
		assert.Len(t, hash, 12)

		hash2, err := RandomHash()
		assert.NoError(t, err)
		assert.NotEqual(t, hash, hash2)
	})

	t.Run("MapChecksum", func(t *testing.T) {
		map1 := map[string]any{
			"user": map[string]string{"name": "Alice"},
			"tags": []string{"admin", "tech"},
			"id":   42,
		}
		map2 := map[string]any{
			"id":   42,
			"tags": []string{"admin", "tech"},
			"user": map[string]string{"name": "Alice"},
		}
		map3 := map[string]any{
			"id":   42,
			"tags": []string{"admin", "tech1"},
			"user": map[string]string{"name": "Alice"},
		}

		sum1, err := MapChecksum(map1)
		assert.NoError(t, err)
		sum2, err := MapChecksum(map2)
		assert.NoError(t, err)
		sum3, err := MapChecksum(map3)
		assert.NoError(t, err)

		assert.Equal(t, sum1, sum2)
		assert.NotEqual(t, sum1, sum3)
		assert.Len(t, sum1, 64)

		empty, err := MapChecksum(map[string]any{})
		assert.NoError(t, err)
		assert.Len(t, empty, 64)

		withNil, err := MapChecksum(map[string]any{"a": nil})
		assert.NoError(t, err)
		assert.NotEqual(t, empty, withNil)

		nested, err := MapChecksum(map[string]any{
			"meta": map[string]any{"count": 1, "ok": true},
		})
		assert.NoError(t, err)
		assert.Len(t, nested, 64)
	})
}
