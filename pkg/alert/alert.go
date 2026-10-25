// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"strings"
)

// Event is the workspace event body sent to an integration destination.
type Event struct {
	Id           string            `json:"id"`
	EventId      string            `json:"eventId"`
	Event        string            `json:"event"`
	WorkspaceId  string            `json:"workspaceId"`
	ActorId      string            `json:"actorId,omitempty"`
	ActorName    string            `json:"actorName,omitempty"`
	ActorType    string            `json:"actorType,omitempty"`
	ResourceType string            `json:"resourceType,omitempty"`
	ResourceId   string            `json:"resourceId,omitempty"`
	IPAddress    string            `json:"ipAddress,omitempty"`
	UserAgent    string            `json:"userAgent,omitempty"`
	OccurredAt   string            `json:"occurredAt"`
	Data         map[string]string `json:"data"`
}

// Sender delivers an event to one integration destination.
type Sender interface {
	Send(ctx context.Context, e Event) error
}

// Title is the one-line headline senders show for an event.
func Title(e Event) string {
	agent := e.Data["agentName"]
	instance := e.Data["instanceId"]

	if e.Data["to"] != "" {
		state := map[string]string{
			"passing":  "recovered",
			"warning":  "degraded",
			"critical": "unhealthy",
			"unknown":  "status unknown",
		}[e.Data["to"]]

		title := fmt.Sprintf(
			"%s %s: %s %s on %s",
			agent,
			state,
			e.Data["checkId"],
			e.Data["to"],
			instance,
		)

		output := []rune(strings.TrimSpace(e.Data["output"]))
		if len(output) > 80 {
			output = append(output[:79], '…')
		}

		if len(output) > 0 && e.Data["to"] != "passing" {
			title += fmt.Sprintf(" (%s)", string(output))
		}

		return title
	}

	endpoint := fmt.Sprintf(
		"%s://%s:%s%s",
		e.Data["scheme"],
		e.Data["address"],
		e.Data["port"],
		e.Data["path"],
	)

	switch e.Data["action"] {
	case "registered":
		return fmt.Sprintf("%s (%s) registered at %s", instance, agent, endpoint)
	case "updated":
		return fmt.Sprintf("%s (%s) moved to %s", instance, agent, endpoint)
	case "deregistered":
		return fmt.Sprintf("%s (%s) deregistered", instance, agent)
	}

	return e.Event
}
