// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/middleware"
)

// requestWithUser attaches an authenticated user to the request context,
// the same way the auth middleware does before a handler runs.
func requestWithUser(r *http.Request, user *db.User) *http.Request {
	return r.WithContext(middleware.WithUserContext(r.Context(), user))
}
