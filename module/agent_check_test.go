// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitUpsertAgentCheckValidation(t *testing.T) {
	// Validation fails before any repository is used, so none are needed.
	check := &Check{}

	cases := []struct {
		name    string
		checkId string
		req     *UpsertAgentCheckRequest
		err     error
	}{
		{"invalid check id", "-bad", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP}, ErrInvalidCheckId},
		{"reserved lease id", db.HealthCheckIDTTL, &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL, TTL: 30}, ErrReservedCheckId},
		{"http needs an absolute path", "http", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "healthz"}, ErrInvalidCheckPath},
		{"http path without spaces", "http", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "/a b"}, ErrInvalidCheckPath},
		{"http scheme", "http", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "/healthz", Scheme: "ftp"}, ErrInvalidCheckScheme},
		{"timeout shorter than interval", "tcp", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP, Interval: 5, Timeout: 5}, ErrInvalidCheckTiming},
		{"default timeout shorter than interval", "tcp", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP, Interval: 2}, ErrInvalidCheckTiming},
		{"ttl requires a ttl", "heartbeat", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL}, ErrInvalidCheckTiming},
		{"unsupported type", "card", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeCard}, ErrUnsupportedCheckType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := check.UpsertAgentCheck("workspace", "support", tc.checkId, tc.req)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
