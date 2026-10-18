// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package slack

// Config holds a customer's Slack credentials.
type Config struct {
	Token      string
	Channel    string
	WebhookURL string
}
