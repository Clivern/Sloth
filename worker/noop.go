// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/broker"

	"github.com/rs/zerolog/log"
)

// HandleNoop accepts a placeholder job and marks it complete.
func (h *handlers) HandleNoop(_ context.Context, msg *broker.Msg) error {
	var payload map[string]string
	err := json.Unmarshal(msg.Data, &payload)
	if err != nil {
		return ErrInvalidPayload
	}

	taskId := db.Id(payload["taskId"])
	if taskId == "" {
		log.Info().
			Str("subject", msg.Subject).
			Msg("Noop job received")
		return nil
	}

	err = h.tasks.MarkRunning(taskId)
	if err != nil {
		return err
	}

	log.Info().
		Str("task_id", taskId.String()).
		Msg("Noop job completed")

	return h.tasks.Complete(taskId, "")
}
