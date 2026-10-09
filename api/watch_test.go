// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"testing"
)

// watchNotImplemented is the skip reason for the sample watch tests.
// Remove the skip in each test when WatchCatalogAction is implemented.
const watchNotImplemented = "watch is not implemented yet (see api/watch.go)"

// TestUnitWatchCatalogAction holds sample cases for GET /watch/catalog from docs/api.md.
func TestUnitWatchCatalogAction(t *testing.T) {
	t.Run("returns changes after the given index", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// GET /watch/catalog?index=42
		// expect: 200 with the catalog changes after index 42 and the latest index
	})

	t.Run("blocks until a change arrives when nothing is newer", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// GET /watch/catalog?index=<latest>
		// expect: the request waits, then returns once a card is published or removed
	})

	t.Run("returns after the wait timeout with the same index", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// GET /watch/catalog?index=<latest> with no changes
		// expect: 200 after the timeout, index unchanged, empty change list
	})

	t.Run("index 0 returns the current catalog", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// GET /watch/catalog?index=0
		// expect: every current agent in the workspace, not a blocking wait
	})

	t.Run("invalid index returns 400", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// GET /watch/catalog?index=abc
		// expect: 400 with a validation error message
	})

	t.Run("only changes from the request workspace are returned", func(t *testing.T) {
		t.Skip(watchNotImplemented)

		// two workspaces publish agents; GET /watch/catalog in ws-1
		// expect: no changes from ws-2
	})
}
