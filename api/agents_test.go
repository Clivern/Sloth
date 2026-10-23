// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAgentsActions(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}

	// instance is a fake agent instance that answers every call with a working task.
	instance := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"task":{"id":"task-1","status":{"state":"TASK_STATE_WORKING"}}}`))
	}))
	defer instance.Close()

	host, port, _ := net.SplitHostPort(instance.Listener.Addr().String())
	portNumber, _ := strconv.Atoi(port)
	live := &db.AgentInstance{Id: "instance-1", AgentId: "agent-1", InstanceId: "directory-1", Address: host, Port: portNumber, Scheme: "http"}

	type setup struct {
		agent     *db.Agent
		intention *db.Intention
		instances []*db.AgentInstance
		task      *db.AgentTask
		skill     *db.AgentSkill
		callable  []*db.CallableAgent
	}

	newRouter := func(s setup) *chi.Mux {
		api := &API{Agents: module.NewAgents(
			&fakeAgentCatalog{agent: s.agent},
			&fakeInstanceRecords{list: s.instances, instance: live},
			&fakeIntentions{match: s.intention},
			&fakeAgentProtections{},
			&fakeTrafficCalls{},
			&fakeAgentTasks{task: s.task},
			&fakeAgentSkills{skill: s.skill, agents: s.callable},
		)}

		router := chi.NewRouter()
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(module.WithAgent(r.Context(), caller)))
			})
		})
		router.Post("/api/v1/workspaces/{workspaceId}/agents/{agentName}/message:send", api.SendAgentMessageAction)
		router.Post("/api/v1/workspaces/{workspaceId}/agents/skills/{skill}/message:send", api.SendSkillMessageAction)
		router.Get("/api/v1/workspaces/{workspaceId}/agents/tasks/{taskId}", api.GetTaskAction)

		return router
	}

	directory := &db.Agent{Id: "agent-1", Name: "user-directory"}
	allow := &db.Intention{Action: db.IntentionActionAllow}
	body := `{"message":{"role":"ROLE_USER","parts":[{"data":{"user_id":"u_123"}}]}}`

	cases := []struct {
		name   string
		setup  setup
		method string
		path   string
		status int
		want   string
	}{
		{
			name:   "send to an agent relays the instance reply",
			setup:  setup{agent: directory, intention: allow, instances: []*db.AgentInstance{live}},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/user-directory/message:send",
			status: http.StatusOK,
			want:   "TASK_STATE_WORKING",
		},
		{
			name:   "send to an agent without an allow intention is forbidden",
			setup:  setup{agent: directory, instances: []*db.AgentInstance{live}},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/user-directory/message:send",
			status: http.StatusForbidden,
			want:   "a2a_call_denied",
		},
		{
			name:   "send to an unknown agent is not found",
			setup:  setup{intention: allow},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/ghost/message:send",
			status: http.StatusNotFound,
			want:   "agent_not_found",
		},
		{
			name:   "send to an agent without a live instance is unavailable",
			setup:  setup{agent: directory, intention: allow},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/user-directory/message:send",
			status: http.StatusServiceUnavailable,
			want:   "no_healthy_instance",
		},
		{
			name:   "send to a skill nobody offers is not found",
			setup:  setup{},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/skills/ghost.skill/message:send",
			status: http.StatusNotFound,
			want:   "skill_not_found",
		},
		{
			name: "send to a skill relays the instance reply",
			setup: setup{
				agent:     directory,
				instances: []*db.AgentInstance{live},
				skill:     &db.AgentSkill{SkillId: "user.profile.get"},
				callable:  []*db.CallableAgent{{Name: "user-directory", Instances: 1, PassingInstances: 1}},
			},
			method: http.MethodPost,
			path:   "/api/v1/workspaces/ws-1/agents/skills/user.profile.get/message:send",
			status: http.StatusOK,
			want:   "task-1",
		},
		{
			name:   "get an unknown task is not found",
			setup:  setup{},
			method: http.MethodGet,
			path:   "/api/v1/workspaces/ws-1/agents/tasks/task-1",
			status: http.StatusNotFound,
			want:   "task_not_found",
		},
		{
			name: "get a task relays the instance reply",
			setup: setup{
				agent:     directory,
				intention: allow,
				task:      &db.AgentTask{TaskId: "task-1", AgentId: "agent-1", InstanceId: &live.Id},
			},
			method: http.MethodGet,
			path:   "/api/v1/workspaces/ws-1/agents/tasks/task-1",
			status: http.StatusOK,
			want:   "task-1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			newRouter(tc.setup).ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(body)))

			require.Equal(t, tc.status, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), tc.want)
		})
	}
}
