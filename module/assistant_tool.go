// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"github.com/clivern/sloth/db"

	"github.com/samber/lo"
)

const (
	AssistantToolLimit    = 50
	AssistantSummaryLimit = 1000
)

// AssistantTool is a dashboard action the assistant may run for the user, gated by the user's permission.
type AssistantTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments,omitempty"`
	Permission  string            `json:"-"`
	Run         AssistantToolRun  `json:"-"`
}

// AssistantToolRun runs a tool in a workspace, for the chat's user, with the arguments the model chose.
type AssistantToolRun func(workspaceId, userId db.Id, args map[string]string) (any, error)

// AssistantTools are the modules the dashboard tools read from.
type AssistantTools struct {
	Workspace   *Workspace
	Agent       *Agent
	Instance    *Instance
	Intention   *Intention
	Traffic     *Traffic
	KeyValue    *KeyValue
	Event       *Event
	Integration *Integration
}

// AssistantWorkspaceSummary is the workspace with counters of what it holds, for questions about the workspace.
type AssistantWorkspaceSummary struct {
	Workspace          *WorkspaceResponse `json:"workspace"`
	Agents             int64              `json:"agents"`
	AgentsByHealth     map[string]int     `json:"agentsByHealth"`
	InstancesByStatus  map[string]int     `json:"instancesByStatus"`
	Intentions         int64              `json:"intentions"`
	IntentionsByAction map[string]int     `json:"intentionsByAction"`
	KeyValues          int64              `json:"kvKeys"`
	GatewayCalls       int64              `json:"gatewayCallsRecorded"`
}

// AssistantIntegration is an integration without its URL, site or index
type AssistantIntegration struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Events    []string `json:"events"`
	IsActive  bool     `json:"isActive"`
	HasSecret bool     `json:"hasSecret"`
}

// NewAssistantTools returns the dashboard tools backed by modules.
func NewAssistantTools(
	workspace *Workspace,
	agent *Agent,
	instance *Instance,
	intention *Intention,
	traffic *Traffic,
	keyValue *KeyValue,
	event *Event,
	integration *Integration,
) *AssistantTools {
	return &AssistantTools{
		Workspace:   workspace,
		Agent:       agent,
		Instance:    instance,
		Intention:   intention,
		Traffic:     traffic,
		KeyValue:    keyValue,
		Event:       event,
		Integration: integration,
	}
}

// List returns every dashboard tool the assistant knows.
func (t *AssistantTools) List() []*AssistantTool {
	return []*AssistantTool{
		{
			Name: "get_workspace",
			Description: "Get this workspace's name, id, handle, member count and the user's own role, with counts of " +
				"agents by health, instances by status, intentions by action, KV keys and gateway calls recorded.",
			Permission: CanGetWorkspace,
			Run:        t.WorkspaceSummary,
		},
		{
			Name:        "list_members",
			Description: "List the workspace members with their name, email and role.",
			Permission:  CanListWorkspaceMembers,
			Run:         t.ListMembers,
		},
		{
			Name:        "list_agents",
			Description: "List the agents registered in the workspace with their version and health.",
			Permission:  CanListAgents,
			Run:         t.ListAgents,
		},
		{
			Name:        "get_agent",
			Description: "Get one agent with its card, skills and health.",
			Arguments:   map[string]string{"name": "agent name"},
			Permission:  CanGetAgent,
			Run:         t.GetAgent,
		},
		{
			Name:        "list_instances",
			Description: "List the running instances of an agent with their health checks.",
			Arguments:   map[string]string{"agent": "agent name"},
			Permission:  CanListInstances,
			Run:         t.ListInstances,
		},
		{
			Name:        "list_intentions",
			Description: "List the intentions: which source agent may or may not call which destination agent and skill.",
			Permission:  CanListIntentions,
			Run:         t.ListIntentions,
		},
		{
			Name:        "list_traffic",
			Description: "List the latest calls through the gateway, newest first, with their decision and result.",
			Permission:  CanListTraffic,
			Run:         t.ListTraffic,
		},
		{
			Name:        "list_kv",
			Description: "List the workspace key-value config, optionally under a key prefix.",
			Arguments:   map[string]string{"prefix": "optional key prefix"},
			Permission:  CanListKeyValue,
			Run:         t.ListKeyValue,
		},
		{
			Name:        "list_integrations",
			Description: "List the integrations that deliver workspace events, such as Slack or webhooks, with the events each sends and whether it is active.",
			Permission:  CanListIntegrations,
			Run:         t.ListIntegrations,
		},
		{
			Name:        "list_events",
			Description: "List the latest workspace events such as health changes, catalog changes and denied calls.",
			Permission:  CanListEvents,
			Run:         t.ListEvents,
		},
	}
}

// WorkspaceSummary returns the workspace with counters of its agents, instances, intentions, KV keys and traffic.
func (t *AssistantTools) WorkspaceSummary(workspaceId, userId db.Id, _ map[string]string) (any, error) {
	workspace, err := t.Workspace.GetWorkspace(workspaceId, &db.User{Id: userId})
	if err != nil {
		return nil, err
	}

	agents, err := t.Agent.ListAgents(workspaceId, AssistantSummaryLimit, 0)
	if err != nil {
		return nil, err
	}

	intentions, err := t.Intention.ListIntentions(workspaceId, AssistantSummaryLimit, 0)
	if err != nil {
		return nil, err
	}

	keyValues, err := t.KeyValue.ListKeyValue(workspaceId, "")
	if err != nil {
		return nil, err
	}

	traffic, err := t.Traffic.ListTraffic(workspaceId, 1, 0)
	if err != nil {
		return nil, err
	}

	summary := &AssistantWorkspaceSummary{
		Workspace:          workspace,
		Agents:             agents.Total,
		AgentsByHealth:     map[string]int{},
		InstancesByStatus:  map[string]int{},
		Intentions:         intentions.Total,
		IntentionsByAction: map[string]int{},
		KeyValues:          keyValues.Total,
		GatewayCalls:       traffic.Total,
	}

	for _, agent := range agents.Agents {
		summary.AgentsByHealth[agent.Health]++
		for status, count := range agent.StatusCounts {
			summary.InstancesByStatus[status] += count
		}
	}

	for _, intention := range intentions.Intentions {
		summary.IntentionsByAction[intention.Action]++
	}

	return summary, nil
}

// ListMembers returns the workspace members with their name, email and role.
func (t *AssistantTools) ListMembers(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	return t.Workspace.ListWorkspaceMembers(workspaceId, AssistantToolLimit, 0)
}

// ListAgents returns the workspace agents with their version and health.
func (t *AssistantTools) ListAgents(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	return t.Agent.ListAgents(workspaceId, AssistantToolLimit, 0)
}

// GetAgent returns one agent, named by the name argument, with its card, skills and health.
func (t *AssistantTools) GetAgent(workspaceId, _ db.Id, args map[string]string) (any, error) {
	return t.Agent.GetAgent(workspaceId, args["name"])
}

// ListInstances returns the instances of the agent named by the agent argument with their health checks.
func (t *AssistantTools) ListInstances(workspaceId, _ db.Id, args map[string]string) (any, error) {
	return t.Instance.ListInstances(workspaceId, args["agent"], AssistantToolLimit, 0)
}

// ListIntentions returns the workspace intentions.
func (t *AssistantTools) ListIntentions(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	return t.Intention.ListIntentions(workspaceId, AssistantToolLimit, 0)
}

// ListTraffic returns the latest gateway calls, newest first.
func (t *AssistantTools) ListTraffic(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	return t.Traffic.ListTraffic(workspaceId, AssistantToolLimit, 0)
}

// ListKeyValue returns the workspace keys, under the prefix argument when given.
func (t *AssistantTools) ListKeyValue(workspaceId, _ db.Id, args map[string]string) (any, error) {
	return t.KeyValue.ListKeyValue(workspaceId, args["prefix"])
}

// ListIntegrations returns the workspace integrations without their URL, site or index.
func (t *AssistantTools) ListIntegrations(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	result, err := t.Integration.ListIntegrations(workspaceId, AssistantToolLimit, 0)
	if err != nil {
		return nil, err
	}

	return lo.Map(result.Integrations, func(item *IntegrationResponse, _ int) *AssistantIntegration {
		return &AssistantIntegration{
			Name:      item.Name,
			Type:      item.Type,
			Events:    item.Events,
			IsActive:  item.IsActive,
			HasSecret: item.HasSecret,
		}
	}), nil
}

// ListEvents returns the latest workspace events.
func (t *AssistantTools) ListEvents(workspaceId, _ db.Id, _ map[string]string) (any, error) {
	return t.Event.ListEvents(workspaceId, AssistantToolLimit, 0)
}
