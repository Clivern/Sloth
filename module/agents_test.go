// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seenRequest is what the fake agent instance received.
type seenRequest struct {
	Method  string
	Path    string
	Version string
	APIKey  string
	Body    string
}

func TestUnitAgents(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}
	message := []byte(`{"message":{"role":"ROLE_USER","parts":[{"data":{"user_id":"u_123"}}]}}`)

	// newInstance starts a fake agent instance that answers with reply and records each request.
	newInstance := func(t *testing.T, reply string) (*agentsFixture, *[]seenRequest) {
		var seen []seenRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			seen = append(seen, seenRequest{r.Method, r.URL.Path, r.Header.Get("A2A-Version"), r.Header.Get("X-Agent-Key"), string(body)})
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(reply))
		}))
		t.Cleanup(server.Close)

		host, port, _ := net.SplitHostPort(server.Listener.Addr().String())
		portNumber, _ := strconv.Atoi(port)

		f := newAgentsFixture(host, portNumber)
		f.agentMetas.values["agent-1/"+AgentMetaA2AURL] = "http://" + server.Listener.Addr().String() + "/a2a/v1/"

		return f, &seen
	}

	t.Run("send to an agent forwards the message to the advertised path and records the task", func(t *testing.T) {
		f, seen := newInstance(t, `{"task":{"id":"task-7f3a","status":{"state":"TASK_STATE_WORKING"}}}`)

		resp, err := f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Version: "1.0", Body: message})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.JSONEq(t, `{"task":{"id":"task-7f3a","status":{"state":"TASK_STATE_WORKING"}}}`, string(resp.Body))

		require.Len(t, *seen, 1)
		assert.Equal(t, http.MethodPost, (*seen)[0].Method)
		assert.Equal(t, "/a2a/v1/message:send", (*seen)[0].Path)
		assert.Equal(t, "1.0", (*seen)[0].Version)
		assert.JSONEq(t, string(message), (*seen)[0].Body)

		require.Len(t, f.tasks.saved, 1)
		assert.Equal(t, "task-7f3a", f.tasks.saved[0].TaskId)
		assert.Equal(t, "support-assistant", f.tasks.saved[0].SourceAgent)
		assert.Equal(t, db.Id("instance-1"), *f.tasks.saved[0].InstanceId)

		require.Len(t, f.traffic.list, 1)
		assert.Equal(t, db.TrafficResultCompleted, f.traffic.list[0].Result)
		assert.Equal(t, VerbSendMessage, f.traffic.list[0].Verb)
	})

	t.Run("a direct message reply records no task", func(t *testing.T) {
		f, _ := newInstance(t, `{"message":{"messageId":"m-1","role":"ROLE_AGENT","parts":[]}}`)

		_, err := f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Body: message})
		require.NoError(t, err)
		assert.Empty(t, f.tasks.saved)
	})

	t.Run("send to an agent adds the protection credential", func(t *testing.T) {
		f, seen := newInstance(t, `{}`)
		secret, err := encrypt.Encrypt("agent-secret")
		require.NoError(t, err)
		config, _ := json.Marshal(ProtectionConfig{Header: "X-Agent-Key", SecretEncrypted: secret})
		f.protections.protection = &db.AgentProtection{Type: db.AgentProtectionTypeAPIKey, Config: string(config), IsActive: true}

		_, err = f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Body: message})
		require.NoError(t, err)
		assert.Equal(t, "agent-secret", (*seen)[0].APIKey)
	})

	t.Run("send to an agent is denied without an allow intention", func(t *testing.T) {
		f, seen := newInstance(t, `{}`)
		f.intentions.match = nil

		_, err := f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrCallDenied)
		assert.Empty(t, *seen)
		require.Len(t, f.traffic.list, 1)
		assert.Equal(t, db.TrafficDecisionDeny, f.traffic.list[0].Decision)
	})

	t.Run("send to an unknown agent is not found", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)
		f.agents.agent = nil

		_, err := f.module.SendMessage(context.Background(), "ws-1", caller, "ghost", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("send to an agent without a live instance fails", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)
		f.instances.live = nil

		_, err := f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
		assert.Equal(t, db.TrafficResultNoInstance, f.traffic.list[0].Result)
	})

	t.Run("send to an instance that cannot be reached fails", func(t *testing.T) {
		f := newAgentsFixture("127.0.0.1", 1)

		_, err := f.module.SendMessage(context.Background(), "ws-1", caller, "user-directory", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrAgentUnreachable)
		assert.Equal(t, db.TrafficResultFailed, f.traffic.list[0].Result)
	})

	t.Run("send to a skill nobody offers is not found", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)

		_, err := f.module.SendSkillMessage(context.Background(), "ws-1", caller, "ghost.skill", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrSkillNotFound)
	})

	t.Run("send to a skill the caller may not call is denied", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{}

		_, err := f.module.SendSkillMessage(context.Background(), "ws-1", caller, "user.profile.get", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrCallDenied)
	})

	t.Run("send to a skill whose agents have no live instance fails", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{{Name: "user-directory", Instances: 1}}

		_, err := f.module.SendSkillMessage(context.Background(), "ws-1", caller, "user.profile.get", &AgentsRequest{Body: message})
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
	})

	t.Run("send to a skill forwards to a live agent and records the skill", func(t *testing.T) {
		f, seen := newInstance(t, `{"task":{"id":"task-1"}}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{{Name: "user-directory", Instances: 1, PassingInstances: 1}}

		_, err := f.module.SendSkillMessage(context.Background(), "ws-1", caller, "user.profile.get", &AgentsRequest{Body: message})
		require.NoError(t, err)
		require.Len(t, *seen, 1)
		assert.Equal(t, "support-assistant", f.skills.agentFilter.Caller)
		assert.Equal(t, "user.profile.get", f.tasks.saved[0].Skill)
		assert.Equal(t, "user.profile.get", f.traffic.list[0].Skill)
	})

	t.Run("get task forwards to the instance that accepted it", func(t *testing.T) {
		f, seen := newInstance(t, `{"id":"task-7f3a","status":{"state":"TASK_STATE_COMPLETED"}}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1", InstanceId: &f.instances.instance.Id}

		resp, err := f.module.GetTask(context.Background(), "ws-1", caller, "task-7f3a", &AgentsRequest{Version: "1.0"})
		require.NoError(t, err)
		assert.Contains(t, string(resp.Body), "TASK_STATE_COMPLETED")
		require.Len(t, *seen, 1)
		assert.Equal(t, http.MethodGet, (*seen)[0].Method)
		assert.Equal(t, "/a2a/v1/tasks/task-7f3a", (*seen)[0].Path)
		assert.Equal(t, VerbGetTask, f.traffic.list[0].Verb)
	})

	t.Run("get task of another caller or an unknown task is not found", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)

		_, err := f.module.GetTask(context.Background(), "ws-1", caller, "task-7f3a", &AgentsRequest{})
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})

	t.Run("get task is denied once the intention is revoked", func(t *testing.T) {
		f, seen := newInstance(t, `{}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1", InstanceId: &f.instances.instance.Id}
		f.intentions.match = &db.Intention{Action: db.IntentionActionDeny}

		_, err := f.module.GetTask(context.Background(), "ws-1", caller, "task-7f3a", &AgentsRequest{})
		assert.ErrorIs(t, err, ErrCallDenied)
		assert.Empty(t, *seen)
	})

	t.Run("get task whose instance is gone fails", func(t *testing.T) {
		f, _ := newInstance(t, `{}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1"}

		_, err := f.module.GetTask(context.Background(), "ws-1", caller, "task-7f3a", &AgentsRequest{})
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
	})
}
