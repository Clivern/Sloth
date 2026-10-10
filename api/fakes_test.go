// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/middleware"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
)

type fakeAccessKeys struct {
	db.AccessKeyRepository
	key       *db.AccessKey
	keys      []*db.AccessKey
	total     int64
	createErr error
	listErr   error
	deleteErr error
}

func (f *fakeAccessKeys) Create(accessKey *db.AccessKey) error {
	return f.createErr
}

func (f *fakeAccessKeys) GetById(id db.Id) (*db.AccessKey, error) {
	return f.key, nil
}

func (f *fakeAccessKeys) GetByKey(key string) (*db.AccessKey, error) {
	return f.key, nil
}

func (f *fakeAccessKeys) DeleteExpired() (int64, error) {
	return 0, nil
}

func (f *fakeAccessKeys) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeAccessKeys) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.AccessKey, error) {
	return f.keys, f.listErr
}

func (f *fakeAccessKeys) Delete(id db.Id) error {
	return f.deleteErr
}

func newAccessKeyAPI(workspace *db.Workspace, keys *fakeAccessKeys) *API {
	return &API{
		Access: module.NewAccess(keys, &fakeWorkspaceRecords{workspace: workspace}),
	}
}

func newAccessKeyRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/api/v1/workspaces/{workspaceId}/access-keys", api.CreateAccessKeyAction)
	router.Get("/api/v1/workspaces/{workspaceId}/access-keys", api.ListAccessKeysAction)
	router.Get("/api/v1/workspaces/{workspaceId}/access-keys/{keyId}", api.GetAccessKeyAction)
	router.Delete("/api/v1/workspaces/{workspaceId}/access-keys/{keyId}", api.DeleteAccessKeyAction)
	return router
}

type fakeAgentCatalog struct {
	db.AgentRepository
	agent     *db.Agent
	list      []*db.Agent
	total     int64
	getErr    error
	listErr   error
	createErr error
	updateErr error
	deleteErr error
}

func (f *fakeAgentCatalog) GetByWorkspaceAndName(workspaceId db.Id, name string) (*db.Agent, error) {
	if f.agent == nil || f.agent.Name != name {
		return nil, f.getErr
	}
	return f.agent, f.getErr
}

func (f *fakeAgentCatalog) GetById(id db.Id) (*db.Agent, error) {
	return f.agent, nil
}

func (f *fakeAgentCatalog) Create(agent *db.Agent) error {
	return f.createErr
}

func (f *fakeAgentCatalog) Update(agent *db.Agent) error {
	return f.updateErr
}

func (f *fakeAgentCatalog) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeAgentCatalog) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.Agent, error) {
	return f.list, f.listErr
}

func (f *fakeAgentCatalog) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.listErr
}

func newAgentAPI(workspace *db.Workspace, agents *fakeAgentCatalog) *API {
	return &API{
		Agent: module.NewAgent(
			agents,
			&fakeInstanceRecords{},
			nil,
			&fakeWorkspaceRecords{workspace: workspace},
		),
	}
}

func newAgentRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/agents", api.ListAgentsAction)
	router.Get("/api/v1/workspaces/{workspaceId}/agents/{agentName}", api.GetAgentAction)
	router.Put("/api/v1/workspaces/{workspaceId}/agents/{agentName}", api.UpsertAgentAction)
	router.Delete("/api/v1/workspaces/{workspaceId}/agents/{agentName}", api.DeleteAgentAction)
	return router
}

type fakeAPIKeys struct {
	db.APIKeyRepository
	apiKey    *db.APIKey
	keys      []*db.APIKey
	total     int64
	getErr    error
	createErr error
	listErr   error
	deleteErr error
}

func (f *fakeAPIKeys) Create(apiKey *db.APIKey) error {
	return f.createErr
}

func (f *fakeAPIKeys) GetByKey(key string) (*db.APIKey, error) {
	return f.apiKey, f.getErr
}

func (f *fakeAPIKeys) GetById(id db.Id) (*db.APIKey, error) {
	return f.apiKey, nil
}

func (f *fakeAPIKeys) DeleteExpired() (int64, error) {
	return 0, nil
}

func (f *fakeAPIKeys) CountByUserId(userId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeAPIKeys) ListByUserId(userId db.Id, limit, offset int) ([]*db.APIKey, error) {
	return f.keys, f.listErr
}

func (f *fakeAPIKeys) Delete(id db.Id) error {
	return f.deleteErr
}

func newAPIKeyRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/api/v1/apiKeys", api.CreateUserAPIKeyAction)
	router.Get("/api/v1/apiKeys", api.ListUserAPIKeysAction)
	router.Get("/api/v1/apiKeys/{apiKeyId}", api.GetUserAPIKeyAction)
	router.Delete("/api/v1/apiKeys/{apiKeyId}", api.DeleteUserAPIKeyAction)
	return router
}

type fakeAuditEvents struct {
	db.AuditEventRepository
	event   *db.AuditEvent
	events  []*db.AuditEvent
	total   int64
	getErr  error
	listErr error
}

func (f *fakeAuditEvents) GetById(id db.Id) (*db.AuditEvent, error) {
	return f.event, f.getErr
}

func (f *fakeAuditEvents) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeAuditEvents) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.AuditEvent, error) {
	return f.events, f.listErr
}

type fakeSubscriptionRecords struct {
	db.SubscriptionRepository
	subscription *db.Subscription
	err          error
}

func (f *fakeSubscriptionRecords) GetByWorkspaceId(workspaceId db.Id) (*db.Subscription, error) {
	return f.subscription, f.err
}

type fakeUsageRecords struct {
	db.UsageRepository
}

func (f *fakeUsageRecords) GetQuantityByPeriod(workspaceId db.Id, utype string, pstart time.Time) (int64, error) {
	return 0, nil
}

type fakeAgentChecks struct {
	db.AgentCheckRepository
	template  *db.AgentCheck
	list      []*db.AgentCheck
	total     int64
	getErr    error
	listErr   error
	createErr error
	updateErr error
	deleteErr error
}

func (f *fakeAgentChecks) GetByAgentAndCheckId(agentId db.Id, checkId string) (*db.AgentCheck, error) {
	return f.template, f.getErr
}

func (f *fakeAgentChecks) Create(check *db.AgentCheck) error {
	return f.createErr
}

func (f *fakeAgentChecks) Update(check *db.AgentCheck) error {
	return f.updateErr
}

func (f *fakeAgentChecks) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeAgentChecks) CountByAgentId(agentId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeAgentChecks) ListPageByAgentId(agentId db.Id, limit, offset int) ([]*db.AgentCheck, error) {
	return f.list, f.listErr
}

func (f *fakeAgentChecks) ListByAgentId(agentId db.Id) ([]*db.AgentCheck, error) {
	return f.list, f.listErr
}

type fakeCheckInstances struct {
	db.AgentInstanceRepository
	instance *db.AgentInstance
}

func (f *fakeCheckInstances) GetByAgentAndInstanceId(agentId db.Id, instanceId string) (*db.AgentInstance, error) {
	return f.instance, nil
}

func (f *fakeCheckInstances) ListByAgentId(agentId db.Id) ([]*db.AgentInstance, error) {
	return nil, nil
}

type fakeHealthChecks struct {
	db.HealthCheckRepository
	check     *db.HealthCheck
	getErr    error
	reportErr error
}

func (f *fakeHealthChecks) GetByInstanceAndCheckId(instanceId db.Id, checkId string) (*db.HealthCheck, error) {
	return f.check, f.getErr
}

func (f *fakeHealthChecks) Report(id db.Id, status, output string, ttlExpiresAt *time.Time) error {
	return f.reportErr
}

func (f *fakeHealthChecks) DeleteByAgentAndCheckId(agentId db.Id, checkId string) error {
	return nil
}

func newCheckAPI(agent *fakeAgentCatalog, checks *fakeAgentChecks, instance *fakeCheckInstances, health *fakeHealthChecks) *API {
	return &API{
		Check: module.NewCheck(
			agent,
			checks,
			instance,
			health,
		),
	}
}

func newCheckRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/agents/{agentName}/checks", api.ListAgentChecksAction)
	router.Put("/api/v1/workspaces/{workspaceId}/agents/{agentName}/checks/{checkId}", api.UpsertAgentCheckAction)
	router.Delete("/api/v1/workspaces/{workspaceId}/agents/{agentName}/checks/{checkId}", api.DeleteAgentCheckAction)
	router.Post("/api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/pass", api.PassCheckAction)
	router.Post("/api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/warn", api.WarnCheckAction)
	router.Post("/api/v1/workspaces/{workspaceId}/agents/{agentName}/instances/{instanceId}/checks/{checkId}/fail", api.FailCheckAction)
	return router
}

type fakeInstanceRecords struct {
	db.AgentInstanceRepository
	instance  *db.AgentInstance
	list      []*db.AgentInstance
	getErr    error
	listErr   error
	createErr error
	updateErr error
	deleteErr error
}

func (f *fakeInstanceRecords) GetByAgentAndInstanceId(agentId db.Id, instanceId string) (*db.AgentInstance, error) {
	return f.instance, f.getErr
}

func (f *fakeInstanceRecords) Create(instance *db.AgentInstance) error {
	return f.createErr
}

func (f *fakeInstanceRecords) Update(instance *db.AgentInstance) error {
	return f.updateErr
}

func (f *fakeInstanceRecords) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeInstanceRecords) ListByAgentId(agentId db.Id) ([]*db.AgentInstance, error) {
	return f.list, f.listErr
}

func (f *fakeInstanceRecords) ListLiveByAgentId(agentId db.Id, now time.Time) ([]*db.AgentInstance, error) {
	return f.list, f.listErr
}

func (f *fakeInstanceRecords) CountByAgentId(agentId db.Id) (int64, error) {
	return int64(len(f.list)), f.listErr
}

func (f *fakeInstanceRecords) ListPageByAgentId(agentId db.Id, limit, offset int) ([]*db.AgentInstance, error) {
	return f.list, f.listErr
}

func (f *fakeInstanceRecords) CountLiveByAgentId(agentId db.Id, now time.Time) (int64, error) {
	return int64(len(f.list)), f.listErr
}

func (f *fakeInstanceRecords) ListLivePageByAgentId(agentId db.Id, now time.Time, limit, offset int) ([]*db.AgentInstance, error) {
	return f.list, f.listErr
}

type fakeInstanceChecks struct {
	db.HealthCheckRepository
	check     *db.HealthCheck
	checks    []*db.HealthCheck
	getErr    error
	listErr   error
	createErr error
	updateErr error
	passErr   error
}

func (f *fakeInstanceChecks) GetByInstanceAndCheckId(instanceId db.Id, checkId string) (*db.HealthCheck, error) {
	return f.check, f.getErr
}

func (f *fakeInstanceChecks) ListByInstanceId(instanceId db.Id) ([]*db.HealthCheck, error) {
	return f.checks, f.listErr
}

func (f *fakeInstanceChecks) Create(check *db.HealthCheck) error {
	return f.createErr
}

func (f *fakeInstanceChecks) Update(check *db.HealthCheck) error {
	return f.updateErr
}

func (f *fakeInstanceChecks) Pass(id db.Id, output string, ttlExpiresAt *time.Time) error {
	return f.passErr
}

func (f *fakeInstanceChecks) Delete(id db.Id) error {
	return nil
}

func newInstanceAPI(workspace *db.Workspace, agent *fakeAgentCatalog, instances *fakeInstanceRecords, checks *fakeInstanceChecks) *API {
	return &API{
		Instance: module.NewInstance(
			agent,
			instances,
			checks,
			&fakeAgentChecks{},
			&fakeWorkspaceRecords{workspace: workspace},
		),
	}
}

func newInstanceRouter(api *API) *chi.Mux {
	base := "/api/v1/workspaces/{workspaceId}/agents/{agentName}/instances"
	router := chi.NewRouter()
	router.Get(base, api.ListInstancesAction)
	router.Get(base+"/{instanceId}", api.GetInstanceAction)
	router.Put(base+"/{instanceId}", api.RegisterInstanceAction)
	router.Delete(base+"/{instanceId}", api.DeregisterInstanceAction)
	return router
}

type fakeIntentions struct {
	db.IntentionRepository
	intention *db.Intention
	match     *db.Intention
	list      []*db.Intention
	total     int64
	getErr    error
	listErr   error
	matchErr  error
	createErr error
	updateErr error
	deleteErr error
}

func (f *fakeIntentions) Create(intention *db.Intention) error {
	return f.createErr
}

func (f *fakeIntentions) GetById(id db.Id) (*db.Intention, error) {
	return f.intention, f.getErr
}

func (f *fakeIntentions) Update(intention *db.Intention) error {
	return f.updateErr
}

func (f *fakeIntentions) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeIntentions) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeIntentions) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.Intention, error) {
	return f.list, f.listErr
}

func (f *fakeIntentions) Match(workspaceId db.Id, source, destination, skill string) (*db.Intention, error) {
	return f.match, f.matchErr
}

func newIntentionAPI(workspace *db.Workspace, intentions *fakeIntentions) *API {
	return &API{
		Intention: module.NewIntention(intentions, &fakeWorkspaceRecords{workspace: workspace}),
	}
}

func newIntentionRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/intentions", api.ListIntentionsAction)
	router.Put("/api/v1/workspaces/{workspaceId}/intentions", api.CreateIntentionAction)
	router.Post("/api/v1/workspaces/{workspaceId}/intentions/check", api.CheckIntentionAction)
	router.Get("/api/v1/workspaces/{workspaceId}/intentions/{intentionId}", api.GetIntentionAction)
	router.Put("/api/v1/workspaces/{workspaceId}/intentions/{intentionId}", api.UpdateIntentionAction)
	router.Delete("/api/v1/workspaces/{workspaceId}/intentions/{intentionId}", api.DeleteIntentionAction)
	return router
}

type fakeInvites struct {
	db.UserInviteRepository
	invite    *db.UserInvite
	list      []*db.UserInvite
	pending   int64
	getErr    error
	listErr   error
	createErr error
	deleteErr error
}

func (f *fakeInvites) Create(invite *db.UserInvite) error {
	return f.createErr
}

func (f *fakeInvites) GetById(id db.Id) (*db.UserInvite, error) {
	return f.invite, f.getErr
}

func (f *fakeInvites) MarkExpiredAsExpired() (int64, error) {
	return 0, nil
}

func (f *fakeInvites) CountPendingByEmailInWorkspace(workspaceId db.Id, email string) (int64, error) {
	return f.pending, nil
}

func (f *fakeInvites) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.UserInvite, error) {
	return f.list, f.listErr
}

func (f *fakeInvites) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return int64(len(f.list)), f.listErr
}

func (f *fakeInvites) Delete(id db.Id) error {
	return f.deleteErr
}

type fakeInviteMailer struct {
	err error
}

func (f *fakeInviteMailer) SendInviteEmail(to, inviteLink, pname string) error {
	return f.err
}

type fakeKeyValues struct {
	db.WorkspaceKeyValueRepository
	item      *db.WorkspaceKeyValue
	items     []*db.WorkspaceKeyValue
	getErr    error
	upsertErr error
	deleteErr error
}

func (f *fakeKeyValues) Upsert(item *db.WorkspaceKeyValue) error {
	return f.upsertErr
}

func (f *fakeKeyValues) Get(workspaceId db.Id, key string) (*db.WorkspaceKeyValue, error) {
	return f.item, f.getErr
}

func (f *fakeKeyValues) Delete(workspaceId db.Id, key string) error {
	return f.deleteErr
}

func (f *fakeKeyValues) ListByPrefix(workspaceId db.Id, prefix string) ([]*db.WorkspaceKeyValue, error) {
	return f.items, nil
}

func newKeyValueRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/kv", api.ListKeyValueAction)
	router.Get("/api/v1/workspaces/{workspaceId}/kv/*", api.GetKeyValueAction)
	router.Put("/api/v1/workspaces/{workspaceId}/kv/*", api.PutKeyValueAction)
	router.Delete("/api/v1/workspaces/{workspaceId}/kv/*", api.DeleteKeyValueAction)
	return router
}

func newKeyValueAPI(workspace *db.Workspace, items *fakeKeyValues) *API {
	return &API{
		KeyValue: module.NewKeyValue(items, &fakeWorkspaceRecords{workspace: workspace}),
	}
}

type fakeUsers struct {
	db.UserRepository
	user      *db.User
	getErr    error
	updateErr error
}

func (f *fakeUsers) GetById(id db.Id) (*db.User, error) {
	return f.user, f.getErr
}

func (f *fakeUsers) Update(user *db.User) error {
	return f.updateErr
}

func newProfileRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/action/profile", api.GetProfileAction)
	router.Put("/api/v1/action/profile", api.UpdateProfileAction)
	return router
}

type fakeConfigs struct {
	db.ConfigRepository
	getErr    error
	updateErr error
}

func (f *fakeConfigs) Get(key string) (*db.Config, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &db.Config{Key: key, Value: "value-of-" + key}, nil
}

func (f *fakeConfigs) Update(key, value string) error {
	return f.updateErr
}

type fakeSessions struct {
	db.SessionRepository
	deleteErr error
}

func (f *fakeSessions) DeleteExpired() (int64, error) {
	return 0, nil
}

func (f *fakeSessions) DeleteByUserId(userId db.Id) error {
	return f.deleteErr
}

type fakeSetupConfigs struct {
	db.ConfigRepository
	installed bool
	createErr error
}

func (f *fakeSetupConfigs) Get(key string) (*db.Config, error) {
	if f.installed {
		return &db.Config{Key: key, Value: "1"}, nil
	}
	return nil, nil
}

func (f *fakeSetupConfigs) Create(key, value string) error {
	return f.createErr
}

type fakeSetupUsers struct {
	db.UserRepository
	createErr error
}

func (f *fakeSetupUsers) Create(user *db.User) error {
	return f.createErr
}

type fakeWorkspaceStats struct {
	db.WorkspaceStatsRepository
	err error
}

func (f *fakeWorkspaceStats) GetByWorkspaceId(workspaceId db.Id) (*db.WorkspaceStats, error) {
	return &db.WorkspaceStats{}, f.err
}

type fakeTrafficCalls struct {
	db.TrafficRepository
	call    *db.TrafficCall
	calls   []*db.TrafficCall
	total   int64
	getErr  error
	listErr error
}

func (f *fakeTrafficCalls) GetById(id db.Id) (*db.TrafficCall, error) {
	return f.call, f.getErr
}

func (f *fakeTrafficCalls) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeTrafficCalls) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.TrafficCall, error) {
	return f.calls, f.listErr
}

func newTrafficRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/traffic", api.ListTrafficAction)
	router.Get("/api/v1/workspaces/{workspaceId}/traffic/{trafficId}", api.GetTrafficAction)
	return router
}

type fakeWorkspaceRecords struct {
	db.WorkspaceRepository
	workspace  *db.Workspace
	membership *db.WorkspaceUser
	list       []*db.Workspace
	total      int64
	getErr     error
	listErr    error
	createErr  error
	updateErr  error
	deleteErr  error
}

func (f *fakeWorkspaceRecords) GetById(id db.Id) (*db.Workspace, error) {
	return f.workspace, f.getErr
}

func (f *fakeWorkspaceRecords) GetByHandle(handle string) (*db.Workspace, error) {
	return nil, nil
}

func (f *fakeWorkspaceRecords) Create(workspace *db.Workspace) error {
	return f.createErr
}

func (f *fakeWorkspaceRecords) Update(workspace *db.Workspace) error {
	return f.updateErr
}

func (f *fakeWorkspaceRecords) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeWorkspaceRecords) List(limit, offset int, userId db.Id) ([]*db.Workspace, error) {
	return f.list, f.listErr
}

func (f *fakeWorkspaceRecords) Count(userId db.Id) (int64, error) {
	return f.total, f.listErr
}

func (f *fakeWorkspaceRecords) GetWorkspaceMembership(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	return f.membership, nil
}

type fakeWorkspaceMembers struct {
	db.WorkspaceUserRepository
	member    *db.WorkspaceUser
	list      []*db.WorkspaceUser
	count     int64
	err       error
	updateErr error
	deleteErr error
}

func (f *fakeWorkspaceMembers) Create(member *db.WorkspaceUser) error {
	return nil
}

func (f *fakeWorkspaceMembers) GetByWorkspaceAndUser(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	return f.member, f.err
}

func (f *fakeWorkspaceMembers) Update(member *db.WorkspaceUser) error {
	return f.updateErr
}

func (f *fakeWorkspaceMembers) Delete(id db.Id) error {
	return f.deleteErr
}

func (f *fakeWorkspaceMembers) ListByWorkspaceId(workspaceId db.Id, limit, offset int) ([]*db.WorkspaceUser, error) {
	return f.list, f.err
}

func (f *fakeWorkspaceMembers) CountByWorkspaceId(workspaceId db.Id) (int64, error) {
	return f.count, f.err
}

func newWorkspaceAPI(records *fakeWorkspaceRecords) *API {
	return &API{
		Workspace: module.NewWorkspace(
			records,
			&fakeWorkspaceMembers{count: 2},
			&fakeSubscriptionRecords{},
			nil,
		),
	}
}

func newWorkspaceRouter(api *API) *chi.Mux {
	router := chi.NewRouter()
	router.Post("/api/v1/workspaces", api.CreateWorkspaceAction)
	router.Get("/api/v1/workspaces", api.ListWorkspacesAction)
	router.Get("/api/v1/workspaces/{workspaceId}", api.GetWorkspaceAction)
	router.Put("/api/v1/workspaces/{workspaceId}", api.UpdateWorkspaceAction)
	router.Delete("/api/v1/workspaces/{workspaceId}", api.DeleteWorkspaceAction)
	return router
}

func (f *fakeUsers) GetByEmail(email string) (*db.User, error) {
	return f.user, nil
}

func (f *fakeSubscriptionRecords) Create(subscription *db.Subscription) error {
	return nil
}

func requestWithUser(r *http.Request, user *db.User) *http.Request {
	return r.WithContext(middleware.WithUserContext(r.Context(), user))
}
