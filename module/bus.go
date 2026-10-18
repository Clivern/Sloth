// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"fmt"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/broker"

	"github.com/rs/zerolog/log"
)

var bus *broker.Client

// StartBus connects the shared NATS client used to publish work.
func StartBus() error {
	client, err := broker.New()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to connect NATS bus")
		return fmt.Errorf("connect nats bus: %w", err)
	}

	bus = client

	log.Info().
		Str("url", client.Config().NATS.URL).
		Str("name", client.Config().NATS.Name).
		Msg("NATS bus connected")

	return nil
}

// GetBus returns the shared NATS publisher.
func GetBus() *broker.Client {
	return bus
}

// StopBus drains and closes the shared NATS client.
func StopBus() {
	err := bus.Conn().Drain()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error draining NATS bus")
		bus.Close()
		return
	}

	bus = nil

	log.Info().
		Msg("NATS bus drained")
}

// EnqueueTask records a pending async task and publishes it on NATS.
func EnqueueTask(taskType string, payload map[string]string, workspaceId db.Id) error {
	taskId, err := db.NewId()
	if err != nil {
		log.Error().
			Err(err).
			Str("type", taskType).
			Msg("Failed to generate task id")
		return err
	}

	payload["taskId"] = taskId.String()
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("type", taskType).
			Str("id", taskId.String()).
			Msg("Failed to encode task payload")
		return err
	}

	body := string(raw)

	err = db.NewAsyncTaskRepository(db.GetDB()).Create(&db.AsyncTask{
		Id:          taskId,
		WorkspaceId: workspaceId,
		Type:        taskType,
		Status:      db.AsyncTaskStatusPending,
		Payload:     &body,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("type", taskType).
			Str("id", taskId.String()).
			Msg("Failed to store task")
		return err
	}

	err = GetBus().Publish(taskType, raw)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("type", taskType).
			Str("id", taskId.String()).
			Msg("Failed to publish task")
		return err
	}

	err = GetBus().Flush()
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("type", taskType).
			Str("id", taskId.String()).
			Msg("Failed to flush NATS bus")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("type", taskType).
		Str("id", taskId.String()).
		Msg("Task enqueued")

	return nil
}

// RepublishPendingTasks publishes pending async tasks to NATS again.
func RepublishPendingTasks() (int, error) {
	tasks, err := db.NewAsyncTaskRepository(db.GetDB()).ListByStatus(db.AsyncTaskStatusPending)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to list pending tasks")
		return 0, err
	}

	for n, task := range tasks {
		err = GetBus().Publish(task.Type, []byte(*task.Payload))
		if err != nil {
			log.Error().
				Err(err).
				Str("id", task.Id.String()).
				Str("type", task.Type).
				Msg("Failed to republish pending task")
			return n, err
		}

		log.Info().
			Str("id", task.Id.String()).
			Str("type", task.Type).
			Msg("Pending task republished")
	}

	err = GetBus().Flush()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to flush NATS bus")
		return 0, err
	}

	return len(tasks), nil
}
