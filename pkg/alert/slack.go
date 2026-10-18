// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/clivern/sloth/pkg/slack"

	"github.com/imroc/req/v3"
	"github.com/samber/lo"
)

// Slack posts events to a Slack incoming webhook.
type Slack struct {
	url  string
	http *req.Client
}

// NewSlack returns a Slack sender for an incoming webhook URL.
func NewSlack(url string) *Slack {
	return &Slack{
		url:  url,
		http: req.C().SetTimeout(10 * time.Second),
	}
}

// Send posts the event name and its data as one Slack message.
func (s *Slack) Send(ctx context.Context, e Event) error {
	keys := lo.Keys(e.Data)
	slices.Sort(keys)

	lines := lo.Map(keys, func(key string, _ int) string {
		return fmt.Sprintf("• *%s*: %s", key, e.Data[key])
	})

	resp, err := s.http.R().
		SetContext(ctx).
		SetBody(slack.Payload{Text: fmt.Sprintf("*%s*\n%s", e.Event, strings.Join(lines, "\n"))}).
		Post(s.url)
	if err != nil {
		return fmt.Errorf("slack request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("slack request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
