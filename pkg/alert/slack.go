// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"

	"github.com/clivern/cognit/pkg/slack"
)

// Slack sends alerts to a Slack channel through an incoming webhook or bot token.
type Slack struct {
	client *slack.Client
}

// NewSlack returns a Slack notifier for a Slack webhook or bot token configuration.
func NewSlack(cfg slack.Config) *Slack {
	return &Slack{
		client: slack.New(cfg),
	}
}

// Send posts the alert as a single Slack message.
func (s *Slack) Send(ctx context.Context, a Alert) error {
	return s.client.Notify(ctx, fmt.Sprintf("[%s] [%s] %s\n%s", a.Severity, a.Status, a.Title, a.Message))
}
