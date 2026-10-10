// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const PagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyConfig holds the Events API v2 routing key of a PagerDuty service.
type PagerDutyConfig struct {
	RoutingKey string
}

// PagerDutyEvent is the Events API v2 request body.
type PagerDutyEvent struct {
	RoutingKey  string           `json:"routing_key"`
	EventAction string           `json:"event_action"`
	DedupKey    string           `json:"dedup_key"`
	Payload     PagerDutyPayload `json:"payload"`
}

// PagerDutyPayload is the incident detail PagerDuty shows for an event.
type PagerDutyPayload struct {
	Summary       string            `json:"summary"`
	Source        string            `json:"source"`
	Severity      string            `json:"severity"`
	CustomDetails map[string]string `json:"custom_details,omitempty"`
}

// PagerDuty sends alerts to a PagerDuty service through the Events API v2.
type PagerDuty struct {
	config   PagerDutyConfig
	endpoint string
	http     *req.Client
}

// NewPagerDuty returns a PagerDuty notifier for a service routing key.
func NewPagerDuty(cfg PagerDutyConfig) *PagerDuty {
	return &PagerDuty{
		config:   cfg,
		endpoint: PagerDutyEventsURL,
		http:     req.C(),
	}
}

// Send triggers or resolves a PagerDuty incident based on the alert status.
func (p *PagerDuty) Send(ctx context.Context, a Alert) error {
	event := PagerDutyEvent{
		RoutingKey:  p.config.RoutingKey,
		EventAction: lo.Ternary(a.Status == StatusResolved, "resolve", "trigger"),
		DedupKey:    a.Key,
		Payload: PagerDutyPayload{
			Summary:       a.Title,
			Source:        a.Source,
			Severity:      a.Severity,
			CustomDetails: lo.Assign(map[string]string{"message": a.Message}, a.Details),
		},
	}

	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(event).
		Post(p.endpoint)
	if err != nil {
		return fmt.Errorf("pagerduty request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("pagerduty request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	log.Info().
		Str("provider", "pagerduty").
		Str("dedup_key", a.Key).
		Msg("Alert sent")

	return nil
}
