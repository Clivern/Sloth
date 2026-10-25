// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

// PagerDutyStatusSeverities maps the status a resource moved to onto an incident severity.
var PagerDutyStatusSeverities = map[string]string{
	"warning":  "warning",
	"critical": "critical",
	"unknown":  "error",
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

// PagerDutyChange is the Change Events API request body.
type PagerDutyChange struct {
	RoutingKey string                 `json:"routing_key"`
	Payload    PagerDutyChangePayload `json:"payload"`
}

// PagerDutyChangePayload is the change PagerDuty shows on the service timeline.
type PagerDutyChangePayload struct {
	Summary       string            `json:"summary"`
	Source        string            `json:"source"`
	Timestamp     string            `json:"timestamp,omitempty"`
	CustomDetails map[string]string `json:"custom_details,omitempty"`
}

// PagerDuty sends events to PagerDuty: info events as changes, the rest as incidents.
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
		endpoint:   fmt.Sprintf("https://%s", site),
		http:       req.C().SetTimeout(10 * time.Second),
	}
}

// Send records an info event as a change and anything else as an incident.
func (p *PagerDuty) Send(ctx context.Context, e Event) error {
	if p.severities[e.Event] == "info" {
		return p.SendInfo(ctx, e)
	}

	return p.SendAlert(ctx, e)
}

// SendInfo records the event as a change on the service timeline, without an incident.
func (p *PagerDuty) SendInfo(ctx context.Context, e Event) error {
	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(PagerDutyChange{
			RoutingKey: p.routingKey,
			Payload: PagerDutyChangePayload{
				Summary:       Title(e),
				Source:        "sloth",
				Timestamp:     e.OccurredAt,
				CustomDetails: e.Data,
			},
		}).
		Post(p.endpoint + "/v2/change/enqueue")

	if err != nil {
		return fmt.Errorf("pagerduty request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("pagerduty request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}

// SendAlert triggers an incident for the event.
func (p *PagerDuty) SendAlert(ctx context.Context, e Event) error {
	incident := PagerDutyEvent{
		RoutingKey:  p.routingKey,
		EventAction: "trigger",
		DedupKey:    e.Id,
		Payload: PagerDutyPayload{
			Summary:       Title(e),
			Source:        "sloth",
			Severity:      p.severities[e.Event],
			CustomDetails: e.Data,
		},
	}

	to := e.Data["to"]
	if to != "" {
		incident.DedupKey = fmt.Sprintf("sloth:%s:%s:%s", e.WorkspaceId, e.ResourceType, e.ResourceId)
		incident.Payload.Severity = PagerDutyStatusSeverities[to]
	}

	if to == "passing" {
		incident.EventAction = "resolve"
		incident.Payload.Severity = "info"
	}

	resp, err := p.http.R().
		SetContext(ctx).
		SetBody(incident).
		Post(p.endpoint + "/v2/enqueue")

	if err != nil {
		return fmt.Errorf("pagerduty request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("pagerduty request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
