// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"errors"
)

var (
	ErrBillingDisabled      = errors.New("stripe billing is not configured")
	ErrInvalidAmount        = errors.New("invalid billing amount")
	ErrWebhookNotConfigured = errors.New("stripe webhook secret is not configured")
)
