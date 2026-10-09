// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/clivern/cognit/api"
	"github.com/clivern/cognit/conf"
	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/middleware"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	cmid "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Setup creates and configures the HTTP server
func SetupServer(Static embed.FS, a *api.API) http.Handler {
	r := chi.NewRouter()

	r.Use(cmid.Recoverer)
	r.Use(cmid.Timeout(
		time.Duration(viper.GetInt("app.timeout")) * time.Second,
	))
	r.Use(middleware.RequestId)
	r.Use(middleware.AppContext)
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestSizeLimit(
		int64(10 * 1024 * 1024),
	))
	r.Use(middleware.Auth())
	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) { // suppress browser favicon requests
		w.WriteHeader(http.StatusNoContent)
	})

	r.Group(func(r chi.Router) { // public — no auth required
		r.Get("/api/v1/public/_health", a.HealthAction)                                   // liveness probe
		r.Get("/api/v1/public/_ready", a.ReadyAction)                                     // readiness probe (db, deps)
		r.Post("/api/v1/public/action/setup", a.SetupAction)                              // initial app setup
		r.Get("/api/v1/public/action/setup/status", a.SetupStatusAction)                  // setup completion status
		r.Post("/api/v1/public/action/logout", a.LogoutAction)                            // user logout
		r.Get("/api/v1/public/action/oauth/github", a.GitHubOAuthStartAction)             // start GitHub OAuth
		r.Get("/api/v1/public/action/oauth/github/callback", a.GitHubOAuthCallbackAction) // GitHub OAuth callback
		r.Get("/api/v1/public/action/oauth/google", a.GoogleOAuthStartAction)             // start Google OAuth
		r.Get("/api/v1/public/action/oauth/google/callback", a.GoogleOAuthCallbackAction) // Google OAuth callback
		if conf.IsSaaS() {
			r.Post("/api/v1/public/action/stripe/webhook", a.StripeWebhookAction) // Stripe billing webhook
		}
	})
	r.Get("/api/v1/me", a.GetMeAction) // current authenticated user
	r.Group(func(r chi.Router) {       // user profile
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Get("/api/v1/action/profile", a.GetProfileAction)    // get user profile
		r.Put("/api/v1/action/profile", a.UpdateProfileAction) // update user profile
	})
	r.Group(func(r chi.Router) { // app settings — admin only
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin}}))
		r.Put("/api/v1/action/settings", a.UpdateSettingsAction) // update app settings
		r.Get("/api/v1/action/settings", a.GetSettingsAction)    // get app settings
	})
	r.Group(func(r chi.Router) { // user API keys
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Post("/api/v1/apiKeys", a.CreateUserAPIKeyAction)              // create user API key
		r.Get("/api/v1/apiKeys", a.ListUserAPIKeysAction)                // list user API keys
		r.Get("/api/v1/apiKeys/{apiKeyId}", a.GetUserAPIKeyAction)       // get user API key
		r.Delete("/api/v1/apiKeys/{apiKeyId}", a.DeleteUserAPIKeyAction) // delete user API key
	})
	r.Group(func(r chi.Router) { // workspaces
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Post("/api/v1/workspaces", a.CreateWorkspaceAction) // create workspace
		r.Get("/api/v1/workspaces", a.ListWorkspacesAction)   // list user workspaces
	})

	r.Route("/api/v1/workspaces/{workspaceId}", func(r chi.Router) { // workspace-scoped resources
		r.Use(middleware.Protect(middleware.Config{Workspace: true}))

		r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetWorkspace})).Get("/", a.GetWorkspaceAction)          // get workspace
		r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpdateWorkspace})).Put("/", a.UpdateWorkspaceAction)    // update workspace
		r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanDeleteWorkspace})).Delete("/", a.DeleteWorkspaceAction) // delete workspace

		r.Route("/invites", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanInviteMember})).Post("/", a.CreateInviteAction)             // create member invite
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListWorkspaceInvites})).Get("/", a.ListInvitesAction)       // list member invites
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetWorkspaceInvite})).Get("/{inviteId}", a.GetInviteAction) // get member invite
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanRemoveInvite})).Delete("/{inviteId}", a.DeleteInviteAction) // revoke member invite
		})

		r.Route("/members", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListWorkspaceMembers})).Get("/", a.ListWorkspaceMembersAction)                // list workspace members
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpdateMemberRole})).Put("/{memberUserId}", a.UpdateWorkspaceMemberRoleAction) // update member role
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanRemoveMember})).Delete("/{memberUserId}", a.DeleteWorkspaceMemberAction)      // remove member
		})

		r.Route("/keys", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanCreateAccessKey})).Post("/", a.CreateAccessKeyAction)          // create workspace access key
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListAccessKeys})).Get("/", a.ListAccessKeysAction)             // list workspace access keys
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetAccessKey})).Get("/{keyId}", a.GetAccessKeyAction)          // get workspace access key
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanDeleteAccessKey})).Delete("/{keyId}", a.DeleteAccessKeyAction) // delete workspace access key
		})

		if conf.IsSaaS() {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspaceBilling})).Get("/billing", a.GetBillingStatusAction)                               // get billing status
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspaceBilling})).Get("/billing/usage", a.GetBillingUsageAction)                          // get billing usage
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpdateWorkspaceBilling})).Post("/billing/checkout", a.CreateBillingCheckoutAction) // start Stripe checkout
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpdateWorkspaceBilling})).Post("/billing/portal", a.CreateBillingPortalAction)     // open Stripe customer portal
		}

		r.Route("/audits", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListWorkspaceAudits})).Get("/", a.ListWorkspaceAuditsAction)      // list audit events
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetWorkspaceAudit})).Get("/{auditId}", a.GetWorkspaceAuditAction) // get audit event
		})

		r.Route("/agents", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListAgents})).Get("/", a.ListAgentsAction)                 // list agents
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpsertAgent})).Put("/{agentName}", a.UpsertAgentAction)    // register or replace an agent card
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetAgent})).Get("/{agentName}", a.GetAgentAction)          // get an agent
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanDeleteAgent})).Delete("/{agentName}", a.DeleteAgentAction) // delete an agent

			r.Route("/{agentName}/instances", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListInstances})).Get("/", a.ListInstancesAction)                       // list agent instances
				r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetInstance})).Get("/{instanceId}", a.GetInstanceAction)               // get an agent instance
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Put("/{instanceId}", a.RegisterInstanceAction)                 // register or update an instance and start its lease
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Put("/{instanceId}/renew", a.RenewInstanceAction)              // renew an instance lease
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Delete("/{instanceId}", a.DeregisterInstanceAction)            // deregister an instance
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Post("/{instanceId}/checks/{checkId}/pass", a.PassCheckAction) // report a ttl check passing
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Post("/{instanceId}/checks/{checkId}/warn", a.WarnCheckAction) // report a ttl check warning
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Post("/{instanceId}/checks/{checkId}/fail", a.FailCheckAction) // report a ttl check critical
			})

			r.Route("/{agentName}/checks", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetAgent})).Get("/", a.ListAgentChecksAction)                 // list health check templates
				r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpsertAgent})).Put("/{checkId}", a.UpsertAgentCheckAction)    // create or replace a health check template
				r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpsertAgent})).Delete("/{checkId}", a.DeleteAgentCheckAction) // delete a health check template
			})
		})

		r.Route("/intentions", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListIntentions})).Get("/", a.ListIntentionsAction)                   // list intentions
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanCreateIntention})).Put("/", a.CreateIntentionAction)                 // create an intention
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanCheckIntention})).Post("/check", a.CheckIntentionAction)             // check whether a call is allowed
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetIntention})).Get("/{intentionId}", a.GetIntentionAction)          // get an intention
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanUpdateIntention})).Put("/{intentionId}", a.UpdateIntentionAction)    // update an intention
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanDeleteIntention})).Delete("/{intentionId}", a.DeleteIntentionAction) // delete an intention
		})

		r.Route("/kv", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListKeyValue})).Get("/", a.ListKeyValueAction)         // list workspace keys
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetKeyValue})).Get("/*", a.GetKeyValueAction)          // get a workspace key
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanPutKeyValue})).Put("/*", a.PutKeyValueAction)          // write a workspace key
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanDeleteKeyValue})).Delete("/*", a.DeleteKeyValueAction) // delete a workspace key
		})

		r.Route("/traffic", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanListTraffic})).Get("/", a.ListTrafficAction)          // list gateway calls
			r.With(middleware.Protect(middleware.Config{User: true, Perm: module.CanGetTraffic})).Get("/{trafficId}", a.GetTrafficAction) // get a gateway call
		})

	})

	r.With(middleware.BasicAuth(viper.GetString("app.metrics.username"), viper.GetString("app.metrics.secret"))).Get("/api/v1/public/_metrics", promhttp.Handler().ServeHTTP) // Prometheus metrics

	dist, err := fs.Sub(Static, "web/dist")
	if err != nil {
		panic(fmt.Sprintf(
			"Error while accessing dist files: %s",
			err.Error(),
		))
	}

	r.Handle("/assets/*", http.StripPrefix("/", http.FileServer(http.FS(dist)))) // static frontend assets
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {                    // SPA fallback — serve index.html
		indexFile, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})

	return r
}

// Run starts the HTTP server with graceful shutdown support
func RunServer(Static embed.FS) error {
	err := db.InitDB(ReadWriteDatabase(), ReadOnlyDatabase()...)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a := api.New()
	handler := SetupServer(Static, a)

	err = module.StartBus()
	if err != nil {
		return fmt.Errorf("failed to start nats bus: %w", err)
	}

	defer func() {
		err := db.CloseDB()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error closing database connection")
		}
	}()

	defer module.StopBus()

	timeout := time.Duration(viper.GetInt("app.timeout")) * time.Second

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		Handler:           handler,
		MaxHeaderBytes:    230 * 1024,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadTimeout:       timeout + 5*time.Second,
		WriteTimeout:      timeout + 5*time.Second,
	}

	serr := make(chan error, 1)

	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Str("edition", conf.Edition()).
			Msg("Starting HTTP server")

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		shutdownTimeout := 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down server")

		err := srv.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		log.Info().
			Msg("Server shutdown complete")
	}

	return nil
}
