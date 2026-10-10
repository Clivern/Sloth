// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package slack

import (
	"context"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const ChatPostMessageURL = "https://slack.com/api/chat.postMessage"

// Client sends Slack notifications for a customer workspace.
type Client struct {
	config Config
	http   *req.Client
}

// Message is a Slack notification payload.
type Message struct {
	Text      string
	Channel   string
	Username  string
	IconEmoji string
}

// Payload is a Slack notification payload.
type Payload struct {
	Text      string `json:"text"`
	Channel   string `json:"channel,omitempty"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// APIResponse is a Slack API response.
type APIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// New returns a Slack client for a customer's workspace.
func New(cfg Config) *Client {
	return &Client{
		config: cfg,
		http:   req.C(),
	}
}

// Notify sends a message using the customer's webhook or bot token.
func (c *Client) Notify(ctx context.Context, text string) error {
	msg := Message{
		Text:    text,
		Channel: c.config.Channel,
	}

	if !lo.IsEmpty(c.config.WebhookURL) {
		return c.PostWebhook(ctx, msg)
	}

	return c.PostMessage(ctx, msg)
}

// PostWebhook posts a message to the customer's Slack incoming webhook.
func (c *Client) PostWebhook(ctx context.Context, msg Message) error {
	resp, err := c.Post(ctx, c.config.WebhookURL, "", msg, nil)
	if err != nil {
		return err
	}

	if resp.String() != "ok" {
		return fmt.Errorf("slack webhook: %s", resp.String())
	}

	log.Info().
		Str("provider", "slack").
		Str("channel", msg.Channel).
		Msg("Slack notification sent")

	return nil
}

// PostMessage posts a message with the customer's bot token via chat.postMessage.
func (c *Client) PostMessage(ctx context.Context, msg Message) error {
	if lo.IsEmpty(msg.Channel) {
		msg.Channel = c.config.Channel
	}

	var result APIResponse

	_, err := c.Post(ctx, ChatPostMessageURL, c.config.Token, msg, &result)
	if err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("slack chat.postMessage: %s", result.Error)
	}

	log.Info().
		Str("provider", "slack").
		Str("channel", msg.Channel).
		Msg("Slack notification sent")

	return nil
}

// Post sends a request to the Slack API.
func (c *Client) Post(ctx context.Context, endpoint, token string, msg Message, dest any) (*req.Response, error) {
	r := c.http.R().
		SetContext(ctx).
		SetBody(Payload{
			Text:      msg.Text,
			Channel:   msg.Channel,
			Username:  msg.Username,
			IconEmoji: msg.IconEmoji,
		})

	if lo.IsNotEmpty(token) {
		r.SetBearerAuthToken(token)
	}

	if !lo.IsNil(dest) {
		r.SetSuccessResult(dest)
	}

	resp, err := r.Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("slack request: %w", err)
	}

	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("slack request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return resp, nil
}
