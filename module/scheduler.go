// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"time"

	"github.com/clivern/sloth/db"

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
func NewScheduler(checks db.HealthCheckRepository, instances db.AgentInstanceRepository, publisher HealthCheckPublisher) *Scheduler {
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
	hold := 2 * time.Minute
	checks, err := s.checks.ClaimDue(now, 100, hold)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to claim due health checks")
		return err
	}

	count := 0
	missed := 0
	for _, check := range checks {
		instance, err := s.instances.GetById(check.InstanceId)
		if err != nil || instance == nil {
			log.Error().Err(err).Str("check_id", check.Id.String()).Msg("Health check instance not found")
			continue
		}

		var def json.RawMessage
		if check.Definition != nil {
			def = json.RawMessage(*check.Definition)
		}

		// A check due again after its interval plus the claim hold
		//  was published before but never recorded.
		var definition CheckDefinition
		if err := json.Unmarshal(def, &definition); err != nil {
			log.Warn().
				Err(err).
				Str("check_id", check.Id.String()).
				Msg("Failed to decode check definition")
		}

		since := check.CreatedAt
		if check.LastRunAt != nil {
			since = *check.LastRunAt
		}

		if now.Sub(since) > time.Duration(definition.Interval)*time.Second+hold {
			missed++
		}

		raw, err := json.Marshal(HealthCheckJob{
			CheckId:    check.Id.String(),
			Type:       check.Type,
			Address:    instance.Address,
			Port:       instance.Port,
			Definition: def,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("check_id", check.Id.String()).
				Msg("Failed to encode health check job")
			return err
		}

		err = s.publisher.Publish(HealthCheckRunSubject, raw)
		if err != nil {
			log.Error().
				Err(err).
				Str("check_id", check.Id.String()).
				Msg("Failed to publish health check job")
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
		log.Error().
			Err(err).
			Msg("Failed to flush health check jobs")
		return err
	}

	if missed > 0 {
		// TODO: Send Dashboard alert.
		// TODO: Send Pageduty alert if system health integration is enabled.
		log.Warn().
			Int("missed", missed).
			Str("subject", HealthCheckRunSubject).
			Msg("Health checks missed their last run, no worker may be consuming")
	}

	if len(checks) > 0 {
		log.Info().
			Int("claimed", len(checks)).
			Int("published", count).
			Msg("Health check jobs scheduled")
	}

	return nil
}
