// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"github.com/clivern/sloth/db"
)

const AssistantToolLimit = 50

// AssistantTool is a dashboard action the assistant may run for the user, gated by the user's permission.
type AssistantTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments,omitempty"`
	Permission  string            `json:"-"`
	Run         AssistantToolRun  `json:"-"`
}

// AssistantToolRun runs a tool in a workspace with the arguments the model chose.
type AssistantToolRun func(workspaceId db.Id, args map[string]string) (any, error)

// AssistantTools are the modules the dashboard tools read from.
type AssistantTools struct {
	Agent     *Agent
	Instance  *Instance
	Intention *Intention
	Traffic   *Traffic
	KeyValue  *KeyValue
	Event     *Event
}

// NewAssistantTools returns the dashboard tools backed by modules.
func NewAssistantTools(agent *Agent, instance *Instance, intention *Intention, traffic *Traffic, keyValue *KeyValue, event *Event) *AssistantTools {
	return &AssistantTools{
		Agent:     agent,
		Instance:  instance,
		Intention: intention,
		Traffic:   traffic,
		KeyValue:  keyValue,
		Event:     event,
	}
}

// List returns every dashboard tool the assistant knows.
func (t *AssistantTools) List() []*AssistantTool {
	return []*AssistantTool{
		{
			Name:        "list_agents",
			Description: "List the agents registered in the workspace with their version and health.",
			Permission:  CanListAgents,
			Run: func(workspaceId db.Id, _ map[string]string) (any, error) {
				return t.Agent.ListAgents(workspaceId, AssistantToolLimit, 0)
			},
		},
		{
			Name:        "get_agent",
			Description: "Get one agent with its card, skills and health.",
			Arguments:   map[string]string{"name": "agent name"},
			Permission:  CanGetAgent,
			Run: func(workspaceId db.Id, args map[string]string) (any, error) {
				return t.Agent.GetAgent(workspaceId, args["name"])
			},
		},
		{
			Name:        "list_instances",
			Description: "List the running instances of an agent with their health checks.",
			Arguments:   map[string]string{"agent": "agent name"},
			Permission:  CanListInstances,
			Run: func(workspaceId db.Id, args map[string]string) (any, error) {
				return t.Instance.ListInstances(workspaceId, args["agent"], AssistantToolLimit, 0)
			},
		},
		{
			Name:        "list_intentions",
			Description: "List the intentions: which source agent may or may not call which destination agent and skill.",
			Permission:  CanListIntentions,
			Run: func(workspaceId db.Id, _ map[string]string) (any, error) {
				return t.Intention.ListIntentions(workspaceId, AssistantToolLimit, 0)
			},
		},
		{
			Name:        "list_traffic",
			Description: "List the latest calls through the gateway, newest first, with their decision and result.",
			Permission:  CanListTraffic,
			Run: func(workspaceId db.Id, _ map[string]string) (any, error) {
				return t.Traffic.ListTraffic(workspaceId, AssistantToolLimit, 0)
			},
		},
		{
			Name:        "list_kv",
			Description: "List the workspace key-value config, optionally under a key prefix.",
			Arguments:   map[string]string{"prefix": "optional key prefix"},
			Permission:  CanListKeyValue,
			Run: func(workspaceId db.Id, args map[string]string) (any, error) {
				return t.KeyValue.ListKeyValue(workspaceId, args["prefix"])
			},
		},
		{
			Name:        "list_events",
			Description: "List the latest workspace events such as health changes, catalog changes and denied calls.",
			Permission:  CanListEvents,
			Run: func(workspaceId db.Id, _ map[string]string) (any, error) {
				return t.Event.ListEvents(workspaceId, AssistantToolLimit, 0)
			},
		},
	}
}
