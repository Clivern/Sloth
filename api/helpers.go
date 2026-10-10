// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/middleware"
	"github.com/clivern/cognit/pkg/util"
)

// ErrorMapping maps a module error to an HTTP status and a translated message.
type ErrorMapping struct {
	Err     error
	Status  int
	Message string
}

// WriteJSON writes data as a JSON response using util.WriteJSON.
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

// WriteError writes the mapped response for err and reports false when err has no mapping.
func (a *API) WriteError(w http.ResponseWriter, err error, mappings []ErrorMapping) bool {
	for _, m := range mappings {
		if errors.Is(err, m.Err) {
			a.WriteJSON(w, m.Status, map[string]any{
				"errorMessage": m.Message,
			})

			return true
		}
	}

	return false
}
