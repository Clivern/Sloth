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

const IncidentAlertEventsURL = "https://api.incident.io/v2/alert_events/http/%s"

// IncidentConfig holds an HTTP alert source secret and its config ID.
type IncidentConfig struct {
	Token               string
	AlertSourceConfigID string
}

// IncidentEvent is the alert event body for an incident.io HTTP alert source.
type IncidentEvent struct {
	Status           string            `json:"status"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	DeduplicationKey string            `json:"deduplication_key"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Incident sends alerts to an incident.io HTTP alert source.
type Incident struct {
	config   IncidentConfig
	endpoint string
	http     *req.Client
}

// NewIncident returns an incident.io notifier for an HTTP alert source.
func NewIncident(cfg IncidentConfig) *Incident {
	return &Incident{
		config:   cfg,
		endpoint: fmt.Sprintf(IncidentAlertEventsURL, cfg.AlertSourceConfigID),
		http:     req.C(),
	}
}

// Send creates or resolves an incident.io alert using the alert status.
func (i *Incident) Send(ctx context.Context, a Alert) error {
	event := IncidentEvent{
		Status:           a.Status,
		Title:            a.Title,
		Description:      a.Message,
		DeduplicationKey: a.Key,
		Metadata:         lo.Assign(map[string]string{"severity": a.Severity, "source": a.Source}, a.Details),
	}

	resp, err := i.http.R().
		SetContext(ctx).
		SetBearerAuthToken(i.config.Token).
		SetBody(event).
		Post(i.endpoint)
	if err != nil {
		return fmt.Errorf("incident.io request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("incident.io request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	log.Info().
		Str("provider", "incident.io").
		Str("dedup_key", a.Key).
		Msg("Alert sent")

	return nil
}
