// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package conf

import (
	"time"
)

const (
	// SlowRequestThreshold is the latency at or above which a request is counted as slow.
	SlowRequestThreshold = time.Second

	// DefaultSessionDuration is the auth session TTL without remember-me.
	DefaultSessionDuration = 7 * 24 * time.Hour

	// RememberMeSessionDuration is the auth session TTL with remember-me.
	RememberMeSessionDuration = 30 * 24 * time.Hour

	// InviteExpiry is how long a workspace invite stays valid.
	InviteExpiry = 7 * 24 * time.Hour
)
