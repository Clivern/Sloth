// Copyright 2026 Sloth. All rights reserved.
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

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/broker"
	"github.com/clivern/sloth/pkg/probe"

	"github.com/rs/zerolog/log"
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

	log.Debug().
		Str("checkId", job.CheckId).
		Str("type", job.Type).
		Str("address", job.Address).
		Int("port", job.Port).
		Str("status", result.Status).
		Msg("Health check probed")

	err = h.checks.Record(
		db.Id(job.CheckId),
		result.Status,
		result.Output,
		time.Now().UTC().Add(interval),
	)
	if err != nil {
		return err
	}

	// A first result that passes is the expected start, not a change.
	if job.Status == result.Status || (job.Status == db.HealthCheckStatusUnknown && result.Status == db.HealthCheckStatusPassing) {
		return nil
	}

	instance, err := h.instances.GetById(db.Id(job.InstanceId))
	if err != nil {
		return err
	}

	agent, err := h.agents.GetById(instance.AgentId)
	if err != nil {
		return err
	}

	h.trigger(agent.WorkspaceId, module.EventAgentHealthChanged, module.EventRequest{
		ActorType:    db.EventActorTypeSystem,
		ResourceType: db.EventResourceTypeHealthCheck,
		ResourceId:   db.Id(job.CheckId),
		Data: map[string]string{
			"agentName":  agent.Name,
			"instanceId": instance.InstanceId,
			"checkId":    job.CheckKey,
			"from":       job.Status,
			"to":         result.Status,
			"output":     result.Output,
		},
	})

	return nil
}
