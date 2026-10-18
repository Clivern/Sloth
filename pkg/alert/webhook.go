// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

// Webhook posts events as JSON signed with an HMAC-SHA256 of the body.
type Webhook struct {
	url    string
	secret string
	http   *req.Client
}

// NewWebhook returns a webhook sender for an endpoint URL and signing secret.
func NewWebhook(url, secret string) *Webhook {
	return &Webhook{
		url:    url,
		secret: secret,
		http:   req.C().SetTimeout(10 * time.Second),
	}
}

// Send posts the event with its signature in the X-Sloth-Signature header.
func (w *Webhook) Send(ctx context.Context, e Event) error {
	body, _ := json.Marshal(e)

	mac := hmac.New(sha256.New, []byte(w.secret))
	mac.Write(body)

	resp, err := w.http.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Sloth-Event", e.Event).
		SetHeader("X-Sloth-Delivery", e.Id).
		SetHeader("X-Sloth-Signature", fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))).
		SetBodyBytes(body).
		Post(w.url)

	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}

	if !resp.IsSuccessState() {
		return fmt.Errorf("webhook request: status %d: %s", resp.StatusCode, resp.Bytes())
	}

	return nil
}
