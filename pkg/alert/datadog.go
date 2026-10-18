// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

// Datadog sends events to the Datadog logs intake of a site.
type Datadog struct {
	apiKey   string
	endpoint string
	http     *req.Client
}

// NewDatadog returns a Datadog sender for a site such as datadoghq.eu and an API key.
func NewDatadog(site, apiKey string) *Datadog {
	return &Datadog{
		apiKey:   apiKey,
		endpoint: fmt.Sprintf("https://http-intake.logs.%s/api/v2/logs", site),
		http:     req.C().SetTimeout(10 * time.Second),
	}
}

// Send ships the event as one log entry tagged with the event and workspace.
func (d *Datadog) Send(ctx context.Context, e Event) error {
	body, _ := json.Marshal(e)

	resp, err := d.http.R().
		SetContext(ctx).
		SetHeader("DD-API-KEY", d.apiKey).
		SetBody([]map[string]string{{
			"ddsource": "sloth",
			"service":  "sloth",
			"ddtags":   fmt.Sprintf("event:%s,workspace:%s", e.Event, e.WorkspaceId),
			"message":  string(body),
		}}).
		Post(d.endpoint)
	if err != nil {
		return fmt.Errorf("datadog request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("datadog request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
