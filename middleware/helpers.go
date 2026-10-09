// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"
)

const (
	// ContextKeyUser is the key for storing user in context
	ContextKeyUser contextKey = "user"
	// ContextKeyAccessKey is the key for storing workspace access key in context
	ContextKeyAccessKey contextKey = "access_key"
	// ContextKeyWorkspace is the key for storing a loaded workspace in context.
	ContextKeyWorkspace contextKey = "workspace"
)

// IsAssetPath returns true if path is a static asset
func IsAssetPath(path string) bool {
	return strings.HasPrefix(path, "/assets/")
}

// ShouldSkipAuth determines if authentication should be skipped for a given path
func ShouldSkipAuth(path string) bool {
	return strings.HasPrefix(path, "/api/v1/public/") || !strings.HasPrefix(path, "/api/v1/")
}

// GetPrincipal returns the authenticated user or workspace access key from context.
func GetPrincipal(r *http.Request) (*db.User, *db.AccessKey, bool) {
	p, ok := module.PrincipalFrom(r.Context())
	if !ok || p == nil {
		return nil, nil, false
	}

	return p.User, p.AccessKey, true
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*db.User, bool) {
	return module.UserFrom(ctx)
}

// GetAccessKeyFromContext retrieves the workspace access key from the request context
func GetAccessKeyFromContext(ctx context.Context) (*db.AccessKey, bool) {
	return module.AccessKeyFrom(ctx)
}

// GetWorkspaceFromContext returns the workspace attached by Protect with Workspace: true.
func GetWorkspaceFromContext(ctx context.Context) (*db.Workspace, bool) {
	workspace, ok := ctx.Value(ContextKeyWorkspace).(*db.Workspace)

	return workspace, ok && workspace != nil
}

// WithUserContext sets the user in the request context
func WithUserContext(ctx context.Context, user *db.User) context.Context {
	ctx = module.WithPrincipal(ctx, &module.Principal{User: user})

	return context.WithValue(ctx, ContextKeyUser, user)
}

// WithAccessKeyContext sets the workspace access key in the request context
func WithAccessKeyContext(ctx context.Context, key *db.AccessKey) context.Context {
	ctx = module.WithPrincipal(ctx, &module.Principal{AccessKey: key})

	return context.WithValue(ctx, ContextKeyAccessKey, key)
}
