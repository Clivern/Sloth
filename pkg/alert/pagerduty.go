// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

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

// PagerDuty triggers PagerDuty incidents through the Events API v2.
type PagerDuty struct {
	routingKey string
	severities map[string]string
	endpoint   string
	http       *req.Client
}

// NewPagerDuty returns a PagerDuty sender for a region host, routing key and event severities.
func NewPagerDuty(site, routingKey string, severities map[string]string) *PagerDuty {
	return &PagerDuty{
		routingKey: routingKey,
		severities: severities,
		endpoint:   fmt.Sprintf("https://%s/v2/enqueue", site),
		http:       req.C().SetTimeout(10 * time.Second),
	}
}

// Send triggers one incident per delivery with the event data as details.
func (p *PagerDuty) Send(ctx context.Context, e Event) error {
	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(PagerDutyEvent{
			RoutingKey:  p.routingKey,
			EventAction: "trigger",
			DedupKey:    e.Id,
			Payload: PagerDutyPayload{
				Summary:       fmt.Sprintf("Sloth %s", e.Event),
				Source:        "sloth",
				Severity:      p.severities[e.Event],
				CustomDetails: e.Data,
			},
		}).
		Post(p.endpoint)

	if err != nil {
		return fmt.Errorf("pagerduty request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("pagerduty request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
