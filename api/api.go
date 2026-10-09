// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/github"
	"github.com/clivern/cognit/pkg/google"
	"github.com/clivern/cognit/pkg/resend"

	"github.com/spf13/viper"
)

// Instance is the process-wide API wired at server startup.
var Instance *API

// API holds shared application modules for HTTP handlers.
type API struct {
	Auth      *module.Auth
	Profile   *module.Profile
	Setup     *module.Setup
	Settings  *module.Settings
	APIKey    *module.APIKey
	Me        *module.Me
	Workspace *module.Workspace
	Invite    *module.Invite
	Access    *module.Access
	Billing   *module.Billing
	Stats     *module.Stats
	Audit     *module.Audit
	Agent     *module.Agent
	Instance  *module.Instance
	Check     *module.Check
	Intention *module.Intention
	KeyValue  *module.KeyValue
	Traffic   *module.Traffic

	GitHubOAuth *github.OAuth
	GoogleOAuth *google.OAuth

	WorkspaceUsers db.WorkspaceUserRepository
	Usage          db.UsageRepository
}

// New wires repositories and modules once for the process.
func New() *API {
	conn := db.GetDB()

	users := db.NewUserRepository(conn)
	sessions := db.NewSessionRepository(conn)
	configs := db.NewConfigRepository(conn)
	workspaces := db.NewWorkspaceRepository(conn)
	workspaceUsers := db.NewWorkspaceUserRepository(conn)
	invites := db.NewUserInviteRepository(conn)
	apiKeys := db.NewAPIKeyRepository(conn)
	accessKeys := db.NewAccessKeyRepository(conn)
	subscriptions := db.NewSubscriptionRepository(conn)
	purchases := db.NewTokenPurchaseRepository(conn)
	stats := db.NewWorkspaceStatsRepository(conn)
	audits := db.NewAuditEventRepository(conn)
	agents := db.NewAgentRepository(conn)
	instances := db.NewAgentInstanceRepository(conn)
	healthChecks := db.NewHealthCheckRepository(conn)
	agentChecks := db.NewAgentCheckRepository(conn)
	intentions := db.NewIntentionRepository(conn)
	kv := db.NewWorkspaceKeyValueRepository(conn)
	traffic := db.NewTrafficRepository(conn)
	usage := db.NewUsageRepository(conn)

	a := &API{
		Auth:      module.NewAuth(users, sessions, configs),
		Profile:   module.NewProfile(users),
		Setup:     module.NewSetup(configs, users),
		Settings:  module.NewSettings(configs),
		APIKey:    module.NewAPIKey(apiKeys),
		Me:        module.NewMe(apiKeys, users, accessKeys),
		Workspace: module.NewWorkspace(workspaces, workspaceUsers, subscriptions, users),
		Invite:    module.NewInvite(invites, users, configs, workspaces, workspaceUsers, resend.NewMailer()),
		Access:    module.NewAccess(accessKeys, workspaces),
		Billing:   module.NewBilling(workspaces, subscriptions, purchases, module.Usage{}),
		Stats:     module.NewStats(workspaces, stats),
		Audit:     module.NewAudit(audits, workspaces),
		Agent:     module.NewAgent(agents, instances, healthChecks, workspaces),
		Instance:  module.NewInstance(agents, instances, healthChecks, agentChecks, workspaces),
		Check:     module.NewCheck(agents, agentChecks, instances, healthChecks, workspaces),
		Intention: module.NewIntention(intentions, workspaces),
		KeyValue:  module.NewKeyValue(kv, workspaces),
		Traffic:   module.NewTraffic(traffic, workspaces),
		GitHubOAuth: github.NewOAuth(github.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.github.client_id"),
			ClientSecret: viper.GetString("app.oauth.github.client_secret"),
			RedirectURL:  viper.GetString("app.oauth.github.redirect_url"),
			Scopes:       []string{"read:user", "user:email"},
			AllowSignup:  true,
		}),
		GoogleOAuth: google.NewOAuth(google.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.google.client_id"),
			ClientSecret: viper.GetString("app.oauth.google.client_secret"),
			RedirectURL:  viper.GetString("app.oauth.google.redirect_url"),
			Scopes:       []string{"openid", "email", "profile"},
		}),
		WorkspaceUsers: workspaceUsers,
		Usage:          usage,
	}

	Instance = a

	return a
}
