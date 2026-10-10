// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import "context"

const (
	// StatusFiring means the condition is active and needs attention.
	StatusFiring = "firing"
	// StatusResolved means a previously firing condition has cleared.
	StatusResolved = "resolved"
	// SeverityCritical is for outages that need an immediate response.
	SeverityCritical = "critical"
	// SeverityError is for failures that need prompt attention.
	SeverityError = "error"
	// SeverityWarning is for degraded conditions that may need attention.
	SeverityWarning = "warning"
	// SeverityInfo is for informational events.
	SeverityInfo = "info"
)

// Alert is a provider-neutral alert sent to a notifier.
type Alert struct {
	Key      string
	Status   string
	Severity string
	Source   string
	Title    string
	Message  string
	Details  map[string]string
}

// Notifier delivers an alert to one destination.
type Notifier interface {
	Send(ctx context.Context, a Alert) error
}
