// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/broker"
	"github.com/clivern/cognit/pkg/probe"
)

// HandleHealthCheck probes one health check job and records the result and next run time.
func (h *handlers) HandleHealthCheck(ctx context.Context, msg *broker.Msg) error {
	var job module.HealthCheckJob

	err := json.Unmarshal(msg.Data, &job)
	if err != nil {
		return ErrInvalidPayload
	}

	var definition module.CheckDefinition
	_ = json.Unmarshal(job.Definition, &definition)

	timeout := time.Duration(definition.Timeout) * time.Second
	interval := time.Duration(definition.Interval) * time.Second

	var result probe.Result
	switch job.Type {
	case db.HealthCheckTypeHTTP:
		target := url.URL{
			Scheme: definition.Scheme,
			Host:   net.JoinHostPort(job.Address, strconv.Itoa(job.Port)),
			Path:   definition.Path,
		}
		result = probe.HTTP(ctx, probe.HTTPTarget{URL: target.String(), Timeout: timeout})

	case db.HealthCheckTypeTCP:
		result = probe.TCP(ctx, probe.TCPTarget{Address: job.Address, Port: job.Port, Timeout: timeout})

	default:
		return fmt.Errorf("unsupported health check type: %s", job.Type)
	}

	return h.checks.Record(
		db.Id(job.CheckId),
		result.Status,
		result.Output,
		time.Now().UTC().Add(interval),
	)
}
