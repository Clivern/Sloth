// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// Context key for request Id
type contextKey string

// RequestIdKey is the context key for storing request IDs
const RequestIdKey contextKey = "request_id"

// RequestId middleware adds a unique request Id to each request
func RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get("X-Request-ID")

		if lo.IsEmpty(requestId) {
			requestId = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestId)
		ctx := context.WithValue(r.Context(), RequestIdKey, requestId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestId retrieves the request Id from the context
func GetRequestId(ctx context.Context) string {
	if requestId, ok := ctx.Value(RequestIdKey).(string); ok {
		return requestId
	}

	return ""
}
