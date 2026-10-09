// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"

	"github.com/clivern/cognit/db"
)

// AppCtx carries request-scoped metadata that can be passed through handlers and services.
type AppCtx struct {
	RequestID string
	UserAgent string
	IPAddress string
	Features  map[string]bool
}

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

// HasFeature reports whether the named feature flag is enabled for this request.
func (c *AppCtx) HasFeature(name string) bool {
	if c == nil {
		return false
	}
	return c.Features[name]
}

type appCtxKey struct{}

// WithAppCtx stores AppCtx on a context.
func WithAppCtx(ctx context.Context, ac *AppCtx) context.Context {
	return context.WithValue(ctx, appCtxKey{}, ac)
}

// AppCtxFrom retrieves AppCtx from a context.
func AppCtxFrom(ctx context.Context) (*AppCtx, bool) {
	ac, ok := ctx.Value(appCtxKey{}).(*AppCtx)
	return ac, ok
}

// Principal is the authenticated actor for a request: a user or a workspace access key.
type Principal struct {
	User      *db.User
	AccessKey *db.AccessKey
}

// IsUser reports whether the principal is an authenticated user.
func (p *Principal) IsUser() bool {
	return p != nil && p.User != nil
}

// IsAccessKey reports whether the principal is a workspace access key.
func (p *Principal) IsAccessKey() bool {
	return p != nil && p.AccessKey != nil
}

type principalKey struct{}

// WithPrincipal stores the authenticated principal on a context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom retrieves the authenticated principal from a context.
func PrincipalFrom(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(*Principal)
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
