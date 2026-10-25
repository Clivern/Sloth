// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"errors"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/broker"

	"github.com/rs/zerolog/log"
)

var ErrInvalidPayload = errors.New("invalid worker payload")

// Handler processes an inbound NATS message.
type Handler func(context.Context, *broker.Msg) error

// Registration binds a subject to a handler.
type Registration struct {
	Subject string
	Handler Handler
}

// Registrations is a list of all registered handlers.
var Registrations []Registration

// Dependencies are the services required by worker handlers.
type Dependencies struct {
	Tasks        db.AsyncTaskRepository
	Checks       db.HealthCheckRepository
	Integrations db.IntegrationRepository
	Instances    db.AgentInstanceRepository
	Agents       db.AgentRepository
}

type handlers struct {
	tasks        db.AsyncTaskRepository
	checks       db.HealthCheckRepository
	integrations db.IntegrationRepository
	instances    db.AgentInstanceRepository
	agents       db.AgentRepository
	trigger      module.EventTrigger
}

// Register attaches all worker handlers.
func Register(deps Dependencies) {
	h := &handlers{
		tasks:        deps.Tasks,
		checks:       deps.Checks,
		integrations: deps.Integrations,
		instances:    deps.Instances,
		agents:       deps.Agents,
		trigger:      module.TriggerEvent,
	}

	On(db.AsyncTaskTypeNoop, h.HandleNoop)
	On(module.HealthCheckRunSubject, h.HandleHealthCheck)
	On(module.IntegrationDeliverTaskType, h.HandleIntegrationDelivery)
}

// On registers a queue worker handler for subject.
func On(subject string, handler Handler) {
	Registrations = append(Registrations, Registration{
		Subject: subject,
		Handler: handler,
	})
}

// Bind attaches all registered handlers to the NATS client using the queue group.
func Bind(client *broker.Client, queue string) error {
	for _, reg := range Registrations {
		_, err := client.QueueSubscribe(reg.Subject, queue, func(msg *broker.Msg) {
			err := reg.Handler(context.Background(), msg)
			if err != nil {
				log.Error().
					Err(err).
					Str("subject", msg.Subject).
					Msg("Worker handler failed")
			}
		})
		if err != nil {
			return err
		}

		log.Info().
			Str("subject", reg.Subject).
			Str("queue", queue).
			Msg("Worker subscribed")
	}

	return nil
}
