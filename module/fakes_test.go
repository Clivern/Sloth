// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/stripe"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type fakeWorkspaces struct {
	db.WorkspaceRepository
	workspace  *db.Workspace
	list       []*db.Workspace
	total      int64
	membership *db.WorkspaceUser
	created    []*db.Workspace
	updated    []*db.Workspace
	deleted    []db.Id
}

func (f *fakeWorkspaces) GetById(id db.Id) (*db.Workspace, error) {
	return f.workspace, nil
}

func (f *fakeWorkspaces) GetByHandle(handle string) (*db.Workspace, error) {
	return nil, nil
}

func (f *fakeWorkspaces) Create(workspace *db.Workspace) error {
	workspace.Id = "ws-new"
	f.created = append(f.created, workspace)

	return nil
}

func (f *fakeWorkspaces) Update(workspace *db.Workspace) error {
	f.updated = append(f.updated, workspace)

	return nil
}

func (f *fakeWorkspaces) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return nil
}

func (f *fakeWorkspaces) List(limit, offset int, userId db.Id) ([]*db.Workspace, error) {
	return f.list, nil
}

func (f *fakeWorkspaces) Count(userId db.Id) (int64, error) {
	return f.total, nil
}

func (f *fakeWorkspaces) GetWorkspaceMembership(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	return f.membership, nil
}

type fakeAgents struct {
	db.AgentRepository
	agent   *db.Agent
	list    []*db.Agent
	total   int64
	created []*db.Agent
	updated []*db.Agent
	deleted []db.Id
	cards   []string
}

func (f *fakeAgents) GetById(id db.Id) (*db.Agent, error) {
	return f.agent, nil
}

func (f *fakeAgents) GetByWorkspaceAndName(workspaceId db.Id, name string) (*db.Agent, error) {
	return f.agent, nil
}

func (f *fakeAgents) Create(agent *db.Agent) error {
	f.created = append(f.created, agent)
	return nil
}

func (f *fakeAgents) Update(agent *db.Agent) error {
	f.updated = append(f.updated, agent)
	return nil
}

func (f *fakeAgents) SetCardIfMissing(id db.Id, card, checksum string) (bool, error) {
	f.cards = append(f.cards, card)
	return true, nil
}

func (f *fakeAgents) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeAgents) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.Agent, error) {
	return f.list, nil
}

func (f *fakeAgents) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, nil
}

type fakeAgentChecks struct {
	db.AgentCheckRepository
	template *db.AgentCheck
	list     []*db.AgentCheck
	total    int64
	created  []*db.AgentCheck
	updated  []*db.AgentCheck
	deleted  []db.Id
}

func (f *fakeAgentChecks) GetByAgentAndCheckId(agentId db.Id, checkId string) (*db.AgentCheck, error) {
	return f.template, nil
}

func (f *fakeAgentChecks) Create(check *db.AgentCheck) error {
	f.created = append(f.created, check)
	return nil
}

func (f *fakeAgentChecks) Update(check *db.AgentCheck) error {
	f.updated = append(f.updated, check)
	return nil
}

func (f *fakeAgentChecks) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeAgentChecks) ListByAgentId(agentId db.Id) ([]*db.AgentCheck, error) {
	return f.list, nil
}

func (f *fakeAgentChecks) ListPageByAgentId(agentId db.Id, limit, offset int) ([]*db.AgentCheck, error) {
	return f.list, nil
}

func (f *fakeAgentChecks) CountByAgentId(agentId db.Id) (int64, error) {
	return f.total, nil
}

type fakeAgentIdentities struct {
	db.AgentIdentityRepository
	identity *db.AgentIdentity
	byName   *db.AgentIdentity
	list     []*db.AgentIdentity
	created  []*db.AgentIdentity
	updated  []*db.AgentIdentity
	deleted  []db.Id
}

func (f *fakeAgentIdentities) GetById(id db.Id) (*db.AgentIdentity, error) {
	return f.identity, nil
}

func (f *fakeAgentIdentities) GetByAgentAndName(agentId db.Id, name string) (*db.AgentIdentity, error) {
	return f.byName, nil
}

func (f *fakeAgentIdentities) Create(identity *db.AgentIdentity) error {
	identity.Id = "identity-1"
	f.created = append(f.created, identity)
	return nil
}

func (f *fakeAgentIdentities) Update(identity *db.AgentIdentity) error {
	f.updated = append(f.updated, identity)
	return nil
}

func (f *fakeAgentIdentities) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeAgentIdentities) ListByAgentId(agentId db.Id) ([]*db.AgentIdentity, error) {
	return f.list, nil
}

type fakeAgentProtections struct {
	db.AgentProtectionRepository
	protection *db.AgentProtection
	created    []*db.AgentProtection
	updated    []*db.AgentProtection
	deleted    []db.Id
}

func (f *fakeAgentProtections) GetByAgentId(agentId db.Id) (*db.AgentProtection, error) {
	return f.protection, nil
}

func (f *fakeAgentProtections) Create(protection *db.AgentProtection) error {
	protection.Id = "protection-1"
	f.created = append(f.created, protection)
	return nil
}

func (f *fakeAgentProtections) Update(protection *db.AgentProtection) error {
	f.updated = append(f.updated, protection)
	return nil
}

func (f *fakeAgentProtections) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakeIdentityTokens struct {
	db.AgentIdentityTokenRepository
	created []*db.AgentIdentityToken
}

func (f *fakeIdentityTokens) Create(token *db.AgentIdentityToken) error {
	f.created = append(f.created, token)
	return nil
}

func (f *fakeIdentityTokens) DeleteExpired() (int64, error) {
	return 0, nil
}

type fakeInstances struct {
	db.AgentInstanceRepository
	list     []*db.AgentInstance
	live     []*db.AgentInstance
	instance *db.AgentInstance
	created  []*db.AgentInstance
	updated  []*db.AgentInstance
	deleted  []db.Id
}

func (f *fakeInstances) ListByAgentId(agentId db.Id) ([]*db.AgentInstance, error) {
	return f.list, nil
}

func (f *fakeInstances) ListLiveByAgentId(agentId db.Id, now time.Time) ([]*db.AgentInstance, error) {
	return f.live, nil
}

func (f *fakeInstances) CountByAgentId(agentId db.Id) (int64, error) {
	return int64(len(f.list)), nil
}

func (f *fakeInstances) ListPageByAgentId(agentId db.Id, limit, offset int) ([]*db.AgentInstance, error) {
	return pageInstances(f.list, limit, offset), nil
}

func (f *fakeInstances) CountLiveByAgentId(agentId db.Id, now time.Time) (int64, error) {
	return int64(len(f.live)), nil
}

func (f *fakeInstances) ListLivePageByAgentId(
	agentId db.Id,
	now time.Time,
	limit, offset int,
) ([]*db.AgentInstance, error) {
	return pageInstances(f.live, limit, offset), nil
}

func pageInstances(items []*db.AgentInstance, limit, offset int) []*db.AgentInstance {
	if offset >= len(items) {
		return nil
	}

	end := min(offset+limit, len(items))

	return items[offset:end]
}

func (f *fakeInstances) GetByAgentAndInstanceId(agentId db.Id, instanceId string) (*db.AgentInstance, error) {
	return f.instance, nil
}

func (f *fakeInstances) Create(instance *db.AgentInstance) error {
	instance.Id = "inst-new"
	f.created = append(f.created, instance)
	return nil
}

func (f *fakeInstances) Update(instance *db.AgentInstance) error {
	f.updated = append(f.updated, instance)
	return nil
}

func (f *fakeInstances) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

// errTest is the error fakes return when a test wants a repository failure.
var errTest = errors.New("boom")

// reportCall records the arguments of one HealthCheckRepository.Report call.
type reportCall struct {
	id           db.Id
	status       string
	output       string
	ttlExpiresAt *time.Time
}

type fakeHealthChecks struct {
	db.HealthCheckRepository
	byInstance     map[db.Id]*db.HealthCheck
	checks         map[db.Id][]*db.HealthCheck
	created        []*db.HealthCheck
	updated        []*db.HealthCheck
	deleted        []db.Id
	deletedByAgent []db.Id
	reports        []reportCall
	passed         []reportCall
}

func (f *fakeHealthChecks) GetByInstanceAndCheckId(instanceId db.Id, checkId string) (*db.HealthCheck, error) {
	return f.byInstance[instanceId], nil
}

func (f *fakeHealthChecks) ListByInstanceId(instanceId db.Id) ([]*db.HealthCheck, error) {
	return f.checks[instanceId], nil
}

func (f *fakeHealthChecks) Create(check *db.HealthCheck) error {
	f.created = append(f.created, check)
	return nil
}

func (f *fakeHealthChecks) Update(check *db.HealthCheck) error {
	f.updated = append(f.updated, check)
	return nil
}

func (f *fakeHealthChecks) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeHealthChecks) DeleteByAgentAndCheckId(agentId db.Id, checkId string) error {
	f.deletedByAgent = append(f.deletedByAgent, agentId)
	return nil
}

func (f *fakeHealthChecks) Pass(id db.Id, output string, ttlExpiresAt *time.Time) error {
	f.passed = append(f.passed, reportCall{id: id, output: output, ttlExpiresAt: ttlExpiresAt})
	return nil
}

func (f *fakeHealthChecks) Report(id db.Id, status, output string, ttlExpiresAt *time.Time) error {
	f.reports = append(f.reports, reportCall{id: id, status: status, output: output, ttlExpiresAt: ttlExpiresAt})
	return nil
}

type fakeConfigs struct {
	db.ConfigRepository
	values  map[string]string
	created []string
	updated []string
	err     error
}

func (f *fakeConfigs) Get(key string) (*db.Config, error) {
	value, ok := f.values[key]
	if f.err != nil || !ok {
		return nil, f.err
	}

	return &db.Config{Key: key, Value: value}, nil
}

func (f *fakeConfigs) Create(key, value string) error {
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[key] = value
	f.created = append(f.created, key)

	return f.err
}

func (f *fakeConfigs) Update(key, value string) error {
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[key] = value
	f.updated = append(f.updated, key)

	return f.err
}

type fakeUsers struct {
	db.UserRepository
	user       *db.User
	byEmail    *db.User
	byProvider *db.User
	created    []*db.User
	updated    []*db.User
	lastLogin  []db.Id
	err        error
}

func (f *fakeUsers) GetById(id db.Id) (*db.User, error) {
	return f.user, f.err
}

func (f *fakeUsers) GetByEmail(email string) (*db.User, error) {
	return f.byEmail, f.err
}

func (f *fakeUsers) GetByProvider(provider, providerUserId string) (*db.User, error) {
	return f.byProvider, f.err
}

func (f *fakeUsers) Create(user *db.User) error {
	if user.Id == "" {
		user.Id = "user-new"
	}
	f.created = append(f.created, user)

	return f.err
}

func (f *fakeUsers) Update(user *db.User) error {
	f.updated = append(f.updated, user)

	return f.err
}

func (f *fakeUsers) UpdateLastLogin(id db.Id) error {
	f.lastLogin = append(f.lastLogin, id)

	return f.err
}

type fakeKeyValues struct {
	db.KeyValueRepository
	item     *db.KeyValue
	upserted []*db.KeyValue
	deleted  []string
	expired  int64
	err      error
}

func (f *fakeKeyValues) Get(key string) (*db.KeyValue, error) {
	return f.item, f.err
}

func (f *fakeKeyValues) Upsert(item *db.KeyValue) error {
	f.upserted = append(f.upserted, item)

	return f.err
}

func (f *fakeKeyValues) Delete(key string) error {
	f.deleted = append(f.deleted, key)

	return f.err
}

func (f *fakeKeyValues) DeleteExpired() (int64, error) {
	return f.expired, f.err
}

type fakeWorkspaceStats struct {
	db.WorkspaceStatsRepository
	err error
}

func (f *fakeWorkspaceStats) GetByWorkspaceId(workspaceId db.Id) (*db.WorkspaceStats, error) {
	return &db.WorkspaceStats{}, f.err
}

type checkFixture struct {
	check        *Check
	agentChecks  *fakeAgentChecks
	healthChecks *fakeHealthChecks
	instances    *fakeInstances
}

// newCheckFixture wires a Check module to in-memory fakes; a nil agent simulates a missing one.
func newCheckFixture(agent *db.Agent) *checkFixture {
	f := &checkFixture{
		agentChecks:  &fakeAgentChecks{},
		healthChecks: &fakeHealthChecks{byInstance: map[db.Id]*db.HealthCheck{}},
		instances:    &fakeInstances{},
	}
	f.check = NewCheck(
		&fakeAgents{agent: agent},
		f.agentChecks,
		f.instances,
		f.healthChecks,
	)

	return f
}

func defaultCheckFixture() *checkFixture {
	return newCheckFixture(&db.Agent{Id: "agent-1", Name: "support-assistant"})
}

type identityFixture struct {
	identity   *Identity
	identities *fakeAgentIdentities
	tokens     *fakeIdentityTokens
}

// newIdentityFixture wires an Identity module to in-memory fakes; a nil agent simulates a missing one.
func newIdentityFixture(agent *db.Agent) *identityFixture {
	f := &identityFixture{identities: &fakeAgentIdentities{}, tokens: &fakeIdentityTokens{}}
	f.identity = NewIdentity(&fakeAgents{agent: agent}, f.identities, f.tokens)

	return f
}

func defaultIdentityFixture() *identityFixture {
	return newIdentityFixture(&db.Agent{Id: "agent-1", Name: "support-assistant"})
}

type protectionFixture struct {
	protection  *Protection
	protections *fakeAgentProtections
}

// newProtectionFixture wires a Protection module to in-memory fakes; a nil agent simulates a missing one.
func newProtectionFixture(agent *db.Agent) *protectionFixture {
	f := &protectionFixture{protections: &fakeAgentProtections{}}
	f.protection = NewProtection(&fakeAgents{agent: agent}, f.protections)

	return f
}

func defaultProtectionFixture() *protectionFixture {
	return newProtectionFixture(&db.Agent{Id: "agent-1", Name: "support-assistant"})
}

type agentFixture struct {
	agent        *Agent
	agents       *fakeAgents
	instances    *fakeInstances
	healthChecks *fakeHealthChecks
}

// newAgentFixture wires an Agent module to in-memory fakes; a nil workspace or agent simulates a missing one.
func newAgentFixture(workspace *db.Workspace, agent *db.Agent) *agentFixture {
	f := &agentFixture{
		agents:       &fakeAgents{agent: agent},
		instances:    &fakeInstances{},
		healthChecks: &fakeHealthChecks{checks: map[db.Id][]*db.HealthCheck{}},
	}
	f.agent = NewAgent(
		f.agents,
		f.instances,
		f.healthChecks,
		&fakeWorkspaces{workspace: workspace},
	)

	return f
}

func defaultAgentFixture() *agentFixture {
	return newAgentFixture(
		&db.Workspace{Id: "ws-1"},
		&db.Agent{
			Id:          "agent-1",
			WorkspaceId: "ws-1",
			Name:        "support-assistant",
			Version:     "1.0.0",
			Card:        lo.ToPtr(`{"name":"Support"}`),
			CardPath:    "/.well-known/agent-card.json",
		},
	)
}

type instanceFixture struct {
	instance     *Instance
	agents       *fakeAgents
	instances    *fakeInstances
	healthChecks *fakeHealthChecks
	agentChecks  *fakeAgentChecks
}

// newInstanceFixture wires an Instance module to in-memory fakes; a nil workspace or agent simulates a missing one.
func newInstanceFixture(workspace *db.Workspace, agent *db.Agent) *instanceFixture {
	f := &instanceFixture{
		agents:    &fakeAgents{agent: agent},
		instances: &fakeInstances{},
		healthChecks: &fakeHealthChecks{
			byInstance: map[db.Id]*db.HealthCheck{},
			checks:     map[db.Id][]*db.HealthCheck{},
		},
		agentChecks: &fakeAgentChecks{},
	}
	f.instance = NewInstance(
		f.agents,
		f.instances,
		f.healthChecks,
		f.agentChecks,
		&fakeWorkspaces{workspace: workspace},
	)
	f.instance.FetchCard = func(ctx context.Context, address string, port int, path string) (map[string]any, error) {
		return nil, errors.New("connection refused")
	}

	return f
}

func defaultInstanceFixture() *instanceFixture {
	return newInstanceFixture(
		&db.Workspace{Id: "ws-1"},
		&db.Agent{Id: "agent-1", WorkspaceId: "ws-1", Name: "support-assistant"},
	)
}

type fakeWorkspaceKeyValues struct {
	db.WorkspaceKeyValueRepository
	item     *db.WorkspaceKeyValue
	items    []*db.WorkspaceKeyValue
	upserted []*db.WorkspaceKeyValue
	deleted  []string
	prefixes []string
	err      error
}

func (f *fakeWorkspaceKeyValues) Get(workspaceId db.Id, key string) (*db.WorkspaceKeyValue, error) {
	return f.item, f.err
}

func (f *fakeWorkspaceKeyValues) Upsert(item *db.WorkspaceKeyValue) error {
	f.upserted = append(f.upserted, item)

	return f.err
}

func (f *fakeWorkspaceKeyValues) Delete(workspaceId db.Id, key string) error {
	f.deleted = append(f.deleted, key)

	return f.err
}

func (f *fakeWorkspaceKeyValues) ListByPrefix(workspaceId db.Id, prefix string) ([]*db.WorkspaceKeyValue, error) {
	f.prefixes = append(f.prefixes, prefix)

	return f.items, f.err
}

type fakeAPIKeys struct {
	db.APIKeyRepository
	item    *db.APIKey
	list    []*db.APIKey
	total   int64
	expired int64
	created []*db.APIKey
	deleted []db.Id
	err     error
}

func (f *fakeAPIKeys) GetById(id db.Id) (*db.APIKey, error) {
	return f.item, f.err
}

func (f *fakeAPIKeys) GetByKey(key string) (*db.APIKey, error) {
	return f.item, f.err
}

func (f *fakeAPIKeys) Create(apiKey *db.APIKey) error {
	apiKey.Id = "key-new"
	f.created = append(f.created, apiKey)

	return f.err
}

func (f *fakeAPIKeys) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeAPIKeys) DeleteExpired() (int64, error) {
	return f.expired, f.err
}

func (f *fakeAPIKeys) ListByUserId(userId db.Id, limit, offset int) ([]*db.APIKey, error) {
	return f.list, f.err
}

func (f *fakeAPIKeys) CountByUserId(userId db.Id) (int64, error) {
	return f.total, f.err
}

type fakeAccessKeys struct {
	db.AccessKeyRepository
	item    *db.AccessKey
	list    []*db.AccessKey
	total   int64
	expired int64
	created []*db.AccessKey
	deleted []db.Id
	err     error
}

func (f *fakeAccessKeys) GetById(id db.Id) (*db.AccessKey, error) {
	return f.item, f.err
}

func (f *fakeAccessKeys) GetByKey(key string) (*db.AccessKey, error) {
	return f.item, f.err
}

func (f *fakeAccessKeys) Create(accessKey *db.AccessKey) error {
	accessKey.Id = "access-new"
	f.created = append(f.created, accessKey)

	return f.err
}

func (f *fakeAccessKeys) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeAccessKeys) DeleteExpired() (int64, error) {
	return f.expired, f.err
}

func (f *fakeAccessKeys) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.AccessKey, error) {
	return f.list, f.err
}

func (f *fakeAccessKeys) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

type fakeAuditEvents struct {
	db.AuditEventRepository
	item  *db.AuditEvent
	list  []*db.AuditEvent
	total int64
	err   error
}

func (f *fakeAuditEvents) GetById(id db.Id) (*db.AuditEvent, error) {
	return f.item, f.err
}

func (f *fakeAuditEvents) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.AuditEvent, error) {
	return f.list, f.err
}

func (f *fakeAuditEvents) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

type fakeTrafficCalls struct {
	db.TrafficRepository
	item  *db.TrafficCall
	list  []*db.TrafficCall
	total int64
	err   error
}

func (f *fakeTrafficCalls) GetById(id db.Id) (*db.TrafficCall, error) {
	return f.item, f.err
}

func (f *fakeTrafficCalls) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.TrafficCall, error) {
	return f.list, f.err
}

func (f *fakeTrafficCalls) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

type fakeWorkspaceUsers struct {
	db.WorkspaceUserRepository
	membership *db.WorkspaceUser
	list       []*db.WorkspaceUser
	total      int64
	created    []*db.WorkspaceUser
	updated    []*db.WorkspaceUser
	deleted    []db.Id
	err        error
}

func (f *fakeWorkspaceUsers) GetByWorkspaceAndUser(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	return f.membership, f.err
}

func (f *fakeWorkspaceUsers) Create(member *db.WorkspaceUser) error {
	f.created = append(f.created, member)

	return f.err
}

func (f *fakeWorkspaceUsers) Update(member *db.WorkspaceUser) error {
	f.updated = append(f.updated, member)

	return f.err
}

func (f *fakeWorkspaceUsers) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeWorkspaceUsers) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.WorkspaceUser, error) {
	return f.list, f.err
}

func (f *fakeWorkspaceUsers) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

type fakeSubscriptions struct {
	db.SubscriptionRepository
	subscription *db.Subscription
	created      []*db.Subscription
	updated      []*db.Subscription
	consumed     []int64
	added        []int64
	err          error
}

func (f *fakeSubscriptions) Create(subscription *db.Subscription) error {
	f.created = append(f.created, subscription)

	return f.err
}

func (f *fakeSubscriptions) Update(subscription *db.Subscription) error {
	f.updated = append(f.updated, subscription)

	return f.err
}

func (f *fakeSubscriptions) GetByWorkspaceId(workspaceId db.Id) (*db.Subscription, error) {
	return f.subscription, f.err
}

func (f *fakeSubscriptions) AddTokens(workspaceId db.Id, tokens int64) error {
	f.added = append(f.added, tokens)

	return f.err
}

func (f *fakeSubscriptions) ConsumeTokens(workspaceId db.Id, tokens int64) error {
	f.consumed = append(f.consumed, tokens)

	return f.err
}

type fakeIntentions struct {
	db.IntentionRepository
	item    *db.Intention
	list    []*db.Intention
	total   int64
	match   *db.Intention
	created []*db.Intention
	updated []*db.Intention
	deleted []db.Id
	err     error
}

func (f *fakeIntentions) GetById(id db.Id) (*db.Intention, error) {
	return f.item, f.err
}

func (f *fakeIntentions) Create(intention *db.Intention) error {
	intention.Id = "intention-new"
	f.created = append(f.created, intention)

	return f.err
}

func (f *fakeIntentions) Update(intention *db.Intention) error {
	f.updated = append(f.updated, intention)

	return f.err
}

func (f *fakeIntentions) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeIntentions) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.Intention, error) {
	return f.list, f.err
}

func (f *fakeIntentions) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

func (f *fakeIntentions) Match(workspaceId db.Id, source, destination, skill string) (*db.Intention, error) {
	return f.match, f.err
}

type fakeSessions struct {
	db.SessionRepository
	session     *db.Session
	list        []*db.Session
	created     []*db.Session
	deleted     []db.Id
	deletedUser []db.Id
	expired     int64
	err         error
}

func (f *fakeSessions) Create(session *db.Session) error {
	f.created = append(f.created, session)

	return f.err
}

func (f *fakeSessions) GetByToken(token string) (*db.Session, error) {
	return f.session, f.err
}

func (f *fakeSessions) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeSessions) DeleteByUserId(userId db.Id) error {
	f.deletedUser = append(f.deletedUser, userId)

	return f.err
}

func (f *fakeSessions) GetByUserId(userId db.Id) ([]*db.Session, error) {
	return f.list, f.err
}

func (f *fakeSessions) DeleteExpired() (int64, error) {
	return f.expired, f.err
}

type fakeUsage struct {
	db.UsageRepository
	quantities  map[string]int64
	incremented []int64
	err         error
}

func (f *fakeUsage) GetQuantityByPeriod(workspaceId db.Id, utype string, pstart time.Time) (int64, error) {
	return f.quantities[utype], f.err
}

func (f *fakeUsage) IncrementByPeriod(
	workspaceId db.Id,
	utype string,
	pstart, pend time.Time,
	quantity int64,
	unit string,
) error {
	f.incremented = append(f.incremented, quantity)

	return f.err
}

type fakeUserInvites struct {
	db.UserInviteRepository
	item          *db.UserInvite
	list          []*db.UserInvite
	pending       []*db.UserInvite
	total         int64
	pendingCount  int64
	expired       int64
	created       []*db.UserInvite
	deleted       []db.Id
	statusUpdates []string
	err           error
}

func (f *fakeUserInvites) GetById(id db.Id) (*db.UserInvite, error) {
	return f.item, f.err
}

func (f *fakeUserInvites) Create(invite *db.UserInvite) error {
	invite.Id = "invite-new"
	f.created = append(f.created, invite)

	return f.err
}

func (f *fakeUserInvites) Delete(id db.Id) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeUserInvites) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.UserInvite, error) {
	return f.list, f.err
}

func (f *fakeUserInvites) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.err
}

func (f *fakeUserInvites) ListPendingByEmail(email string) ([]*db.UserInvite, error) {
	return f.pending, f.err
}

func (f *fakeUserInvites) UpdateStatus(id db.Id, status string, acceptedAt *time.Time) error {
	f.statusUpdates = append(f.statusUpdates, status)

	return f.err
}

func (f *fakeUserInvites) MarkExpiredAsExpired() (int64, error) {
	return f.expired, f.err
}

func (f *fakeUserInvites) CountPendingByEmailInWorkspace(workspaceId db.Id, email string) (int64, error) {
	return f.pendingCount, f.err
}

type fakeMailer struct {
	sent []string
	err  error
}

func (f *fakeMailer) SendInviteEmail(to, inviteLink, pname string) error {
	f.sent = append(f.sent, to)

	return f.err
}

type fakeTokenPurchases struct {
	db.TokenPurchaseRepository
	inserted bool
	created  []*db.TokenPurchase
	err      error
}

func (f *fakeTokenPurchases) Create(purchase *db.TokenPurchase) (bool, error) {
	f.created = append(f.created, purchase)

	return f.inserted, f.err
}

// enableBilling switches the process to the SaaS edition with Stripe, reset after the test.
func enableBilling(t *testing.T) *stripe.Client {
	t.Helper()
	t.Setenv("COGNIT_EDITION", "saas")
	viper.Set("app.billing.provider", "stripe")
	viper.Set("app.billing.secret_key", "sk_test_123")
	viper.Set("app.billing.webhook_secret", "whsec_test")
	t.Cleanup(func() {
		viper.Set("app.billing.provider", "")
		viper.Set("app.billing.secret_key", "")
		viper.Set("app.billing.webhook_secret", "")
	})

	client, err := stripe.New()
	require.NoError(t, err)

	return client
}

func newBillingModule(
	workspace *db.Workspace,
	subscription *db.Subscription,
) (*Billing, *fakeSubscriptions, *fakeTokenPurchases) {
	subscriptions := &fakeSubscriptions{subscription: subscription}
	purchases := &fakeTokenPurchases{inserted: true}
	billing := NewBilling(&fakeWorkspaces{workspace: workspace}, subscriptions, purchases, Usage{})

	return billing, subscriptions, purchases
}

func saveIntentionRequest(action string) *SaveIntentionRequest {
	return &SaveIntentionRequest{
		Source:      IntentionSource{Agent: "planner"},
		Destination: IntentionDestination{Agent: "invoice", Skill: "invoice.extract"},
		Action:      action,
	}
}

func newInviteModule(
	invites *fakeUserInvites,
	users *fakeUsers,
	members *fakeWorkspaceUsers,
	mailer *fakeMailer,
) *Invite {
	return NewInvite(
		invites,
		users,
		&fakeConfigs{},
		&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		members,
		mailer,
	)
}

func googleIdentity() *OAuthIdentity {
	return &OAuthIdentity{
		Provider:       "google",
		ProviderUserID: "g-123",
		Email:          "ada@example.com",
		Name:           "Ada",
	}
}

// newOAuthAuth wires an Auth module for OAuth logins.
func newOAuthAuth(users *fakeUsers, sessions *fakeSessions, configs *fakeConfigs) *Auth {
	return NewAuth(users, sessions, configs)
}

// newPerm builds a checker without a database, with an in-memory workspace repository reporting membership.
func newPerm(membership *db.WorkspaceUser) *Perm {
	perm := NewPerm(nil)
	perm.WorkspaceRepository = &fakeWorkspaces{
		membership: membership,
		workspace:  &db.Workspace{Id: "ws-1"},
	}

	return perm
}

func newWorkspaceModule(
	workspaces *fakeWorkspaces,
	members *fakeWorkspaceUsers,
	subscriptions *fakeSubscriptions,
) *Workspace {
	return NewWorkspace(workspaces, members, subscriptions, &fakeUsers{})
}
