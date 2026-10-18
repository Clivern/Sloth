// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/alert"
	"github.com/clivern/sloth/pkg/broker"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// HandleIntegrationDelivery sends one event to one integration and records the task outcome.
func (h *handlers) HandleIntegrationDelivery(ctx context.Context, msg *broker.Msg) error {
	var payload map[string]string
	err := json.Unmarshal(msg.Data, &payload)
	if err != nil {
		return ErrInvalidPayload
	}

	taskId := db.Id(payload["taskId"])

	err = h.tasks.MarkRunning(taskId)
	if err != nil {
		return err
	}

	item, err := h.integrations.GetById(db.Id(payload["integrationId"]))
	if err != nil {
		return err
	}

	if item == nil || !item.IsActive {
		log.Info().
			Str("taskId", taskId.String()).
			Str("integrationId", payload["integrationId"]).
			Msg("Integration gone or disabled, delivery skipped")
		return h.tasks.Complete(taskId, "")
	}

	var config module.IntegrationConfig
	_ = json.Unmarshal([]byte(lo.FromPtr(item.Config)), &config)

	secret := ""
	if config.SecretEncrypted != "" {
		secret, err = encrypt.Decrypt(config.SecretEncrypted)
		if err != nil {
			return h.tasks.Fail(taskId, err.Error())
		}
	}

	event := alert.Event{
		Id:           taskId.String(),
		EventId:      payload["eventId"],
		Event:        payload["event"],
		WorkspaceId:  item.WorkspaceId.String(),
		ActorId:      payload["actorId"],
		ActorName:    payload["actorName"],
		ActorType:    payload["actorType"],
		ResourceType: payload["resourceType"],
		ResourceId:   payload["resourceId"],
		IPAddress:    payload["ipAddress"],
		UserAgent:    payload["userAgent"],
		OccurredAt:   payload["occurredAt"],
	}
	_ = json.Unmarshal([]byte(payload["data"]), &event.Data)

	var sender alert.Sender
	switch item.Type {
	case db.IntegrationTypeWebhook:
		sender = alert.NewWebhook(config.URL, secret)
	case db.IntegrationTypeSlack:
		sender = alert.NewSlack(config.URL)
	case db.IntegrationTypeDatadog:
		sender = alert.NewDatadog(config.Site, secret)
	case db.IntegrationTypeSplunk:
		sender = alert.NewSplunk(config.URL, secret, config.Index)
	case db.IntegrationTypePagerDuty:
		sender = alert.NewPagerDuty(config.Site, secret, module.PagerDutySeverities)
	}

	err = sender.Send(ctx, event)
	if err != nil {
		log.Error().
			Err(err).
			Str("taskId", taskId.String()).
			Str("integrationId", item.Id.String()).
			Str("type", item.Type).
			Str("event", event.Event).
			Msg("Integration delivery failed")
		return h.tasks.Fail(taskId, err.Error())
	}

	log.Info().
		Str("taskId", taskId.String()).
		Str("integrationId", item.Id.String()).
		Str("type", item.Type).
		Str("event", event.Event).
		Msg("Integration delivery sent")

	return h.tasks.Complete(taskId, "")
}
