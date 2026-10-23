// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	// AgentMetaA2AURL is the agent meta key holding the A2A url its card advertised before Sloth rewrote it.
	AgentMetaA2AURL = "a2a_url"
	// AgentsTimeout bounds each call forwarded to an instance.
	AgentsTimeout = 60 * time.Second
	// MaxAgentsResponseSize caps the response body read from an instance.
	MaxAgentsResponseSize = 10 << 20
	// MaxSkillCandidates caps the agents considered when a message is addressed by skill.
	MaxSkillCandidates = 1000

	VerbSendMessage = "SendMessage"
	VerbGetTask     = "GetTask"
)

var (
	ErrCallDenied        = errors.New("call denied")
	ErrNoHealthyInstance = errors.New("no healthy instance")
	ErrTaskNotFound      = errors.New("task not found")
	ErrAgentUnreachable  = errors.New("agent unreachable")
)

// AgentsRequest is the part of an incoming call forwarded to an instance.
type AgentsRequest struct {
	Version string
	Body    []byte
}

// AgentsResponse is the instance reply, relayed to the caller as is.
type AgentsResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// AgentsTrafficMeta is the JSON stored in ai_traffic.meta for a forwarded call.
type AgentsTrafficMeta struct {
	AgentId    db.Id `json:"agentId"`
	InstanceId db.Id `json:"instanceId"`
}

// sendReply is the part of a message:send reply that names a task; a direct message has none.
type sendReply struct {
	Task struct {
		Id string `json:"id"`
	} `json:"task"`
}

// Agents is the module that forwards messages between agents and fetches the tasks they start.
type Agents struct {
	AgentRepository      db.AgentRepository
	AgentMetaRepository  db.AgentMetaRepository
	InstanceRepository   db.AgentInstanceRepository
	IntentionRepository  db.IntentionRepository
	ProtectionRepository db.AgentProtectionRepository
	TrafficRepository    db.TrafficRepository
	TaskRepository       db.AgentTaskRepository
	SkillRepository      db.AgentSkillRepository
	http                 *req.Client
}

// NewAgents creates an agents module with the given repositories.
func NewAgents(
	agents db.AgentRepository,
	agentMetas db.AgentMetaRepository,
	instances db.AgentInstanceRepository,
	intentions db.IntentionRepository,
	protections db.AgentProtectionRepository,
	traffic db.TrafficRepository,
	tasks db.AgentTaskRepository,
	skills db.AgentSkillRepository,
) *Agents {
	return &Agents{
		AgentRepository:      agents,
		AgentMetaRepository:  agentMetas,
		InstanceRepository:   instances,
		IntentionRepository:  intentions,
		ProtectionRepository: protections,
		TrafficRepository:    traffic,
		TaskRepository:       tasks,
		SkillRepository:      skills,
		http:                 req.C().SetTimeout(AgentsTimeout),
	}
}

// AdvertisedA2AURL returns the HTTP+JSON url an agent card advertises, falling back to its first interface or legacy url.
func AdvertisedA2AURL(card map[string]any) string {
	interfaces, _ := card["supportedInterfaces"].([]any)

	var first string
	for _, iface := range interfaces {
		entry, _ := iface.(map[string]any)
		link, _ := entry["url"].(string)
		first = lo.CoalesceOrEmpty(first, link)

		if entry["protocolBinding"] == "HTTP+JSON" && link != "" {
			return link
		}
	}

	legacy, _ := card["url"].(string)

	return lo.CoalesceOrEmpty(first, legacy)
}

// SendMessage sends a message from caller to a named agent; only intentions without a skill apply.
func (a *Agents) SendMessage(ctx context.Context, workspaceId db.Id, caller *db.Agent, agentName string, r *AgentsRequest) (*AgentsResponse, error) {
	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	call := newCall(workspaceId, caller.Name, agent.Name, "", VerbSendMessage)

	allowed, err := a.allowed(workspaceId, caller.Name, agent.Name, "")
	if err != nil {
		return nil, err
	}
	if !allowed {
		a.record(call, db.TrafficResultDenied)
		return nil, ErrCallDenied
	}

	return a.send(ctx, call, caller, agent, r)
}

// SendSkillMessage sends a message from caller to a random live agent it may call for a skill.
func (a *Agents) SendSkillMessage(ctx context.Context, workspaceId db.Id, caller *db.Agent, skill string, r *AgentsRequest) (*AgentsResponse, error) {
	offered, err := a.SkillRepository.GetSkill(workspaceId, skill)
	if err != nil {
		return nil, err
	}
	if offered == nil {
		return nil, ErrSkillNotFound
	}

	call := newCall(workspaceId, caller.Name, "", skill, VerbSendMessage)

	// The same query as the catalog, so the caller can reach exactly the agents it was shown.
	callable, _, err := a.SkillRepository.ListCallableAgents(workspaceId, skill, &db.ListCallableAgentsFilter{
		Caller: caller.Name,
		Limit:  MaxSkillCandidates,
	})
	if err != nil {
		return nil, err
	}
	if len(callable) == 0 {
		a.record(call, db.TrafficResultDenied)
		return nil, ErrCallDenied
	}

	live := lo.Filter(callable, func(agent *db.CallableAgent, _ int) bool { return agent.PassingInstances > 0 })
	if len(live) == 0 {
		a.record(call, db.TrafficResultNoInstance)
		return nil, ErrNoHealthyInstance
	}

	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, lo.Sample(live).Name)
	if err != nil {
		return nil, err
	}

	call.DestinationInstance = agent.Name

	return a.send(ctx, call, caller, agent, r)
}

// GetTask fetches a task caller started from the instance that accepted it.
func (a *Agents) GetTask(ctx context.Context, workspaceId db.Id, caller *db.Agent, taskId string, r *AgentsRequest) (*AgentsResponse, error) {
	// Tasks are looked up by their caller, so another agent's task is not found.
	task, err := a.TaskRepository.GetBySourceAndTaskId(workspaceId, caller.Name, taskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	agent, err := a.AgentRepository.GetById(task.AgentId)
	if err != nil {
		return nil, err
	}

	call := newCall(workspaceId, caller.Name, agent.Name, task.Skill, VerbGetTask)

	// Revoking the intention also stops the caller from reading tasks it already started.
	allowed, err := a.allowed(workspaceId, caller.Name, agent.Name, task.Skill)
	if err != nil {
		return nil, err
	}
	if !allowed {
		a.record(call, db.TrafficResultDenied)
		return nil, ErrCallDenied
	}

	// The task lives only on the instance that accepted it.
	if task.InstanceId == nil {
		a.record(call, db.TrafficResultNoInstance)
		return nil, ErrNoHealthyInstance
	}

	instance, err := a.InstanceRepository.GetById(*task.InstanceId)
	if err != nil {
		return nil, err
	}

	return a.forward(ctx, call, agent, instance, http.MethodGet, "/tasks/"+url.PathEscape(taskId), r)
}

// send forwards a message:send to a random live instance of agent and remembers where a returned task lives.
func (a *Agents) send(ctx context.Context, call *db.TrafficCall, caller, agent *db.Agent, r *AgentsRequest) (*AgentsResponse, error) {
	instances, err := a.InstanceRepository.ListLiveByAgentId(agent.Id, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		a.record(call, db.TrafficResultNoInstance)
		return nil, ErrNoHealthyInstance
	}

	instance := lo.Sample(instances)

	resp, err := a.forward(ctx, call, agent, instance, http.MethodPost, "/message:send", r)
	if err != nil {
		return nil, err
	}

	var reply sendReply
	_ = json.Unmarshal(resp.Body, &reply)

	if reply.Task.Id != "" {
		err = a.TaskRepository.Save(&db.AgentTask{
			WorkspaceId: call.WorkspaceId,
			TaskId:      reply.Task.Id,
			SourceAgent: caller.Name,
			AgentId:     agent.Id,
			InstanceId:  &instance.Id,
			Skill:       call.Skill,
		})
		// The message is already delivered, so a lost record only stops the caller from polling the task.
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", call.WorkspaceId.String()).
				Str("taskId", reply.Task.Id).
				Msg("Failed to record agent task")
		}
	}

	return resp, nil
}

// forward calls one instance, relays its reply and records the call.
func (a *Agents) forward(ctx context.Context, call *db.TrafficCall, agent *db.Agent, instance *db.AgentInstance, method, path string, r *AgentsRequest) (*AgentsResponse, error) {
	target, err := a.target(agent, instance)
	if err != nil {
		return nil, err
	}

	request := a.http.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		DisableAutoReadResponse()

	if r.Version != "" {
		request.SetHeader("A2A-Version", r.Version)
	}

	if len(r.Body) > 0 {
		request.SetContentType("application/json").SetBodyBytes(r.Body)
	}

	err = a.protect(request, agent.Id)
	if err != nil {
		return nil, err
	}

	meta, _ := json.Marshal(AgentsTrafficMeta{AgentId: agent.Id, InstanceId: instance.Id})
	call.DestinationInstance = instance.InstanceId
	call.Meta = lo.ToPtr(string(meta))

	start := time.Now()
	resp, err := request.Send(method, target+path)
	call.LatencyMs = int(time.Since(start).Milliseconds())

	if err != nil {
		var timeout interface{ Timeout() bool }
		a.record(call, lo.Ternary(errors.As(err, &timeout) && timeout.Timeout(), db.TrafficResultTimeout, db.TrafficResultFailed))

		return nil, fmt.Errorf("%w: %v", ErrAgentUnreachable, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxAgentsResponseSize))
	if err != nil {
		a.record(call, db.TrafficResultFailed)
		return nil, fmt.Errorf("%w: %v", ErrAgentUnreachable, err)
	}

	a.record(call, lo.Ternary(resp.IsSuccessState(), db.TrafficResultCompleted, db.TrafficResultFailed))

	return &AgentsResponse{
		StatusCode:  resp.StatusCode,
		ContentType: resp.GetContentType(),
		Body:        body,
	}, nil
}

// target is the instance address with the scheme and path of the url the agent card advertised.
func (a *Agents) target(agent *db.Agent, instance *db.AgentInstance) (string, error) {
	target := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(instance.Address, strconv.Itoa(instance.Port)),
	}

	meta, err := a.AgentMetaRepository.Get(agent.Id, AgentMetaA2AURL)
	if err != nil {
		return "", err
	}
	if meta != nil {
		advertised, _ := url.Parse(meta.Value)
		target.Scheme = lo.CoalesceOrEmpty(advertised.Scheme, target.Scheme)
		target.Path = strings.TrimSuffix(advertised.Path, "/")
	}

	return target.String(), nil
}

// protect adds the credential a protected agent expects from Sloth.
func (a *Agents) protect(request *req.Request, agentId db.Id) error {
	protection, err := a.ProtectionRepository.GetByAgentId(agentId)
	if err != nil {
		return err
	}
	if protection == nil || !protection.IsActive {
		return nil
	}

	var config ProtectionConfig
	_ = json.Unmarshal([]byte(protection.Config), &config)

	secret, err := encrypt.Decrypt(config.SecretEncrypted)
	if err != nil {
		return err
	}

	switch protection.Type {
	case db.AgentProtectionTypeAPIKey:
		request.SetHeader(config.Header, secret)
	case db.AgentProtectionTypeBasicAuth:
		request.SetBasicAuth(config.Username, secret)
	}

	return nil
}

// allowed reports whether an intention lets source call destination for skill; no intention means deny.
func (a *Agents) allowed(workspaceId db.Id, source, destination, skill string) (bool, error) {
	intention, err := a.IntentionRepository.Match(workspaceId, source, destination, skill)
	if err != nil {
		return false, err
	}

	return intention != nil && intention.Action == db.IntentionActionAllow, nil
}

// record stores a call in traffic with its result; a failed write never fails the call.
func (a *Agents) record(call *db.TrafficCall, result string) {
	call.Result = result
	call.Decision = lo.Ternary(result == db.TrafficResultDenied, db.TrafficDecisionDeny, db.TrafficDecisionAllow)

	err := a.TrafficRepository.Create(call)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", call.WorkspaceId.String()).
			Str("verb", call.Verb).
			Msg("Failed to record agent traffic")
	}
}

// newCall starts the traffic record of a call from source.
func newCall(workspaceId db.Id, source, destination, skill, verb string) *db.TrafficCall {
	return &db.TrafficCall{
		WorkspaceId:         workspaceId,
		SourceInstance:      source,
		DestinationInstance: destination,
		Skill:               skill,
		Verb:                verb,
	}
}
