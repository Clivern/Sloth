// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/rs/zerolog/log"
)

// HealthCheckRunSubject is the NATS subject workers consume to run one check.
const HealthCheckRunSubject = "health.check.run"

// HealthCheckPublisher is the part of the NATS bus the scheduler publishes to.
type HealthCheckPublisher interface {
	Publish(subject string, data []byte) error
	Flush() error
}

// HealthCheckJob is one due check with the instance address a worker needs to probe it.
type HealthCheckJob struct {
	CheckId    string          `json:"checkId"`
	Type       string          `json:"type"`
	Address    string          `json:"address"`
	Port       int             `json:"port"`
	Definition json.RawMessage `json:"definition,omitempty"`
}

// Scheduler publishes due health checks to NATS for workers to run.
type Scheduler struct {
	checks    db.HealthCheckRepository
	instances db.AgentInstanceRepository
	publisher HealthCheckPublisher
}

// NewScheduler returns a scheduler that reads due checks from checks and publishes to publisher.
func NewScheduler(
	checks db.HealthCheckRepository,
	instances db.AgentInstanceRepository,
	publisher HealthCheckPublisher,
) *Scheduler {
	return &Scheduler{checks: checks, instances: instances, publisher: publisher}
}

// Run publishes due checks every few seconds until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info().
		Str("subject", HealthCheckRunSubject).
		Msg("Health check scheduler started")

	for {
		select {
		case <-ctx.Done():
			log.Info().
				Msg("Health check scheduler stopped")
			return
		case now := <-ticker.C:
			err := s.Tick(now)
			if err != nil {
				log.Error().Err(err).Msg("Health check scheduler tick failed")
			}
		}
	}
}

// Tick claims the checks due at now and publishes one job for each.
func (s *Scheduler) Tick(now time.Time) error {
	checks, err := s.checks.ClaimDue(now, 100, 2*time.Minute)
	if err != nil {
		return err
	}

	count := 0
	for _, check := range checks {
		instance, err := s.instances.GetById(check.InstanceId)
		if err != nil || instance == nil {
			log.Error().Err(err).Str("check_id", check.Id.String()).Msg("Health check instance not found")
			continue
		}

		var definition json.RawMessage
		if check.Definition != nil {
			definition = json.RawMessage(*check.Definition)
		}

		raw, err := json.Marshal(HealthCheckJob{
			CheckId:    check.Id.String(),
			Type:       check.Type,
			Address:    instance.Address,
			Port:       instance.Port,
			Definition: definition,
		})
		if err != nil {
			return err
		}

		err = s.publisher.Publish(HealthCheckRunSubject, raw)
		if err != nil {
			return err
		}

		count++
		log.Debug().
			Str("check_id", check.Id.String()).
			Str("type", check.Type).
			Msg("Health check job published")
	}

	err = s.publisher.Flush()
	if err != nil {
		return err
	}

	if len(checks) > 0 {
		log.Info().
			Int("claimed", len(checks)).
			Int("published", count).
			Msg("Health check jobs scheduled")
	}

	return nil
}
