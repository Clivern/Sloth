// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/worker"

	"github.com/rs/zerolog/log"
)

// RunWorker starts the NATS worker and blocks until shutdown.
func RunWorker() error {
	err := db.InitDB(ReadWriteDatabase(), ReadOnlyDatabase()...)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	defer func() {
		err := db.CloseDB()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error closing database connection")
		}
	}()

	// Handlers trigger events, which enqueue deliveries on the shared bus.
	err = module.StartBus()
	if err != nil {
		return fmt.Errorf("failed to start nats bus: %w", err)
	}

	defer module.StopBus()

	client := module.GetBus()

	nats := client.Config().NATS

	worker.Register(worker.Dependencies{
		Tasks:        db.NewAsyncTaskRepository(db.GetDB(false)),
		Checks:       db.NewHealthCheckRepository(db.GetDB(true)),
		Integrations: db.NewIntegrationRepository(db.GetDB(true)),
		Instances:    db.NewAgentInstanceRepository(db.GetDB()),
		Agents:       db.NewAgentRepository(db.GetDB()),
	})

	err = worker.Bind(client, nats.Queue)
	if err != nil {
		return fmt.Errorf("failed to bind workers: %w", err)
	}

	log.Info().
		Str("url", nats.URL).
		Str("name", nats.Name).
		Str("queue", nats.Queue).
		Msg("Starting NATS worker")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit

	log.Info().
		Str("signal", sig.String()).
		Msg("Received shutdown signal")

	err = client.Conn().Drain()
	if err != nil {
		return fmt.Errorf("failed to drain nats connection: %w", err)
	}

	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if client.Conn().IsClosed() {
			break
		}

		select {
		case <-deadline:
			return fmt.Errorf("nats drain timed out")
		case <-ticker.C:
		}
	}

	log.Info().
		Msg("Worker shutdown complete")

	return nil
}
