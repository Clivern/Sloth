// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/middleware"
	"github.com/clivern/cognit/pkg/util"
)

// WriteJSON writes a value as a JSON response. It shares util.WriteJSON with
// middleware, which cannot import this package.
func (a *API) WriteJSON(w http.ResponseWriter, statusCode int, data any) error {
	return util.WriteJSON(w, statusCode, data)
}

// DecodeJSON reads and decodes JSON from the request body.
func (a *API) DecodeJSON(r *http.Request, v any) error {
	return util.DecodeJSON(r, v)
}

// DecodeAndValidate decodes JSON and validates the struct in one step.
func (a *API) DecodeAndValidate(r *http.Request, v any) error {
	return util.DecodeAndValidate(r, v)
}

// WriteValidationError writes a decode or validation error as a 400 JSON response.
func (a *API) WriteValidationError(w http.ResponseWriter, err error) {
	util.WriteValidationError(w, err)
}

// ParsePagination parses limit and offset from query parameters.
func (a *API) ParsePagination(r *http.Request) (limit, offset int) {
	return util.ParsePagination(r)
}

// GetUser returns the authenticated user from the request context.
func (a *API) GetUser(r *http.Request) (*db.User, bool) {
	return middleware.GetUserFromContext(r.Context())
}
