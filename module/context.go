// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"

	"github.com/clivern/sloth/db"
)

// AppCtx carries request-scoped metadata that can be passed through handlers and services.
type AppCtx struct {
	RequestID string
	UserAgent string
	IPAddress string
	Features  map[string]bool
}

// Principal is the authenticated actor for a request: a user or a workspace access key.
type Principal struct {
	User      *db.User
	AccessKey *db.AccessKey
}

// AppCtxKey is the context key for AppCtx.
type AppCtxKey struct{}

// PrincipalKey is the context key for the authenticated principal.
type PrincipalKey struct{}

// AgentKey is the context key for the agent authenticated by its identity.
type AgentKey struct{}

// NewAppCtx builds an AppCtx from request metadata.
func NewAppCtx(requestID, userAgent, ipAddress string, features map[string]bool) *AppCtx {
	if features == nil {
		features = map[string]bool{}
	}

	return &AppCtx{
		RequestID: requestID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Features:  features,
	}
}

// HasFeature reports whether the named feature flag is enabled for this request. TODO: remove if not used in future.
func (c *AppCtx) HasFeature(name string) bool {
	if c == nil {
		return false
	}

	return c.Features[name]
}

// WithAppCtx stores AppCtx on a context.
func WithAppCtx(ctx context.Context, ac *AppCtx) context.Context {
	return context.WithValue(ctx, AppCtxKey{}, ac)
}

// AppCtxFrom retrieves AppCtx from a context.
func AppCtxFrom(ctx context.Context) (*AppCtx, bool) {
	ac, ok := ctx.Value(AppCtxKey{}).(*AppCtx)

	return ac, ok
}

// IsUser reports whether the principal is an authenticated user. TODO: remove if not used in future.
func (p *Principal) IsUser() bool {
	return p != nil && p.User != nil
}

// IsAccessKey reports whether the principal is a workspace access key. TODO: remove if not used in future.
func (p *Principal) IsAccessKey() bool {
	return p != nil && p.AccessKey != nil
}

// WithPrincipal stores the authenticated principal on a context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, PrincipalKey{}, p)
}

// PrincipalFrom retrieves the authenticated principal from a context.
func PrincipalFrom(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(PrincipalKey{}).(*Principal)

	return p, ok
}

// UserFrom retrieves the authenticated user from a context.
func UserFrom(ctx context.Context) (*db.User, bool) {
	p, ok := PrincipalFrom(ctx)
	if !ok || p == nil || p.User == nil {
		return nil, false
	}

	return p.User, true
}

// AccessKeyFrom retrieves the workspace access key from a context.
func AccessKeyFrom(ctx context.Context) (*db.AccessKey, bool) {
	p, ok := PrincipalFrom(ctx)
	if !ok || p == nil || p.AccessKey == nil {
		return nil, false
	}

	return p.AccessKey, true
}

// WithAgent stores the agent authenticated by its identity on a context.
func WithAgent(ctx context.Context, agent *db.Agent) context.Context {
	return context.WithValue(ctx, AgentKey{}, agent)
}

// AgentFrom retrieves the agent authenticated by its identity from a context.
func AgentFrom(ctx context.Context) (*db.Agent, bool) {
	agent, ok := ctx.Value(AgentKey{}).(*db.Agent)

	return agent, ok
}
