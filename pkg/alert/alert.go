// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import "context"

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
