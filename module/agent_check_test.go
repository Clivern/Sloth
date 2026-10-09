// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitNormalizeCheck(t *testing.T) {
	t.Run("http fills defaults", func(t *testing.T) {
		definition, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "/healthz"})
		require.NoError(t, err)
		assert.Equal(t, "/healthz", definition.Path)
		assert.Equal(t, "http", definition.Scheme)
		assert.Equal(t, DefaultCheckInterval, definition.Interval)
		assert.Equal(t, DefaultCheckTimeout, definition.Timeout)
	})

	t.Run("http needs an absolute path", func(t *testing.T) {
		_, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "healthz"})
		assert.ErrorIs(t, err, ErrInvalidCheckPath)

		_, err = normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "/a b"})
		assert.ErrorIs(t, err, ErrInvalidCheckPath)
	})

	t.Run("timeout must be shorter than interval", func(t *testing.T) {
		_, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP, Interval: 5, Timeout: 5})
		assert.ErrorIs(t, err, ErrInvalidCheckTiming)
	})

	t.Run("tcp ignores http fields", func(t *testing.T) {
		definition, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP, Path: "/x", Port: 9000})
		require.NoError(t, err)
		assert.Empty(t, definition.Path)
		assert.Equal(t, 9000, definition.Port)
	})

	t.Run("ttl requires a ttl", func(t *testing.T) {
		_, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL})
		assert.ErrorIs(t, err, ErrInvalidCheckTiming)

		definition, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL, TTL: 60, Interval: 10})
		require.NoError(t, err)
		assert.Equal(t, &CheckDefinition{TTL: 60}, definition)
	})

	t.Run("rejects other types", func(t *testing.T) {
		_, err := normalizeCheck(&UpsertAgentCheckRequest{Type: db.HealthCheckTypeCard})
		assert.ErrorIs(t, err, ErrUnsupportedCheckType)
	})
}

func TestUnitInstanceStatus(t *testing.T) {
	now := time.Now().UTC()
	later := now.Add(time.Minute)
	earlier := now.Add(-time.Second)

	lease := func(expires time.Time) *db.HealthCheck {
		return &db.HealthCheck{
			CheckId:      db.HealthCheckIDTTL,
			Type:         db.HealthCheckTypeTTL,
			Source:       db.HealthCheckSourceLease,
			Status:       db.HealthCheckStatusPassing,
			TTLExpiresAt: &expires,
		}
	}
	check := func(checkType, status string) *db.HealthCheck {
		return &db.HealthCheck{
			CheckId: checkType,
			Type:    checkType,
			Source:  db.HealthCheckSourceAgent,
			Status:  status,
		}
	}

	t.Run("live lease alone is passing", func(t *testing.T) {
		assert.Equal(t, db.AgentInstanceStatusPassing, instanceStatus([]*db.HealthCheck{lease(later)}, now))
	})

	t.Run("expired lease is critical", func(t *testing.T) {
		assert.Equal(t, db.AgentInstanceStatusCritical, instanceStatus([]*db.HealthCheck{lease(earlier)}, now))
	})

	t.Run("missing lease is critical", func(t *testing.T) {
		checks := []*db.HealthCheck{check(db.HealthCheckTypeHTTP, db.HealthCheckStatusPassing)}
		assert.Equal(t, db.AgentInstanceStatusCritical, instanceStatus(checks, now))
	})

	t.Run("worst check wins", func(t *testing.T) {
		checks := []*db.HealthCheck{
			lease(later),
			check(db.HealthCheckTypeHTTP, db.HealthCheckStatusPassing),
			check(db.HealthCheckTypeTCP, db.HealthCheckStatusWarning),
		}
		assert.Equal(t, db.AgentInstanceStatusWarning, instanceStatus(checks, now))

		checks = append(checks, check(db.HealthCheckTypeHTTP, db.HealthCheckStatusCritical))
		assert.Equal(t, db.AgentInstanceStatusCritical, instanceStatus(checks, now))
	})

	t.Run("expired template ttl is critical", func(t *testing.T) {
		heartbeat := check(db.HealthCheckTypeTTL, db.HealthCheckStatusPassing)
		heartbeat.TTLExpiresAt = &earlier
		assert.Equal(t, db.HealthCheckStatusCritical, effectiveCheckStatus(heartbeat, now))
		assert.Equal(t, db.AgentInstanceStatusCritical, instanceStatus([]*db.HealthCheck{lease(later), heartbeat}, now))
	})
}

func TestUnitWorstStatus(t *testing.T) {
	assert.Equal(t, db.AgentInstanceStatusCritical, worstStatus(nil))
	assert.Equal(t, db.AgentInstanceStatusPassing, worstStatus([]string{"passing", "passing"}))
	assert.Equal(t, db.AgentInstanceStatusWarning, worstStatus([]string{"passing", "warning"}))
	assert.Equal(t, db.AgentInstanceStatusCritical, worstStatus([]string{"warning", "critical", "passing"}))
}
