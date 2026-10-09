// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"testing"
)

// discoverNotImplemented is the skip reason for the sample discover tests.
// Remove the skip in each test when DiscoverAgentsAction or ResolveDiscoverAction is implemented.
const discoverNotImplemented = "discover is not implemented yet (see api/discover.go)"

// TestUnitDiscoverAgentsAction holds sample cases for GET /discover from docs/api.md.
func TestUnitDiscoverAgentsAction(t *testing.T) {
	t.Run("filters by skill", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?skill=invoice.extract
		// expect: only agents that advertise invoice.extract
	})

	t.Run("filters by tags", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?tags=finance
		// expect: only agents tagged finance
	})

	t.Run("passing=true drops instances that are not passing", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?skill=invoice.extract&passing=true
		// expect: agents whose instances are all critical are excluded
	})

	t.Run("limit caps the number of results", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?skill=invoice.extract&limit=5
		// expect: at most five agents
	})

	t.Run("filters by input mode and capability", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?input_mode=application/pdf&capability=streaming
		// expect: only agents whose cards declare both
	})

	t.Run("unknown skill returns an empty list", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// GET /discover?skill=does.not.exist
		// expect: 200 with an empty list, not 404
	})
}

// TestUnitResolveDiscoverAction holds sample cases for POST /discover/resolve from docs/api.md.
func TestUnitResolveDiscoverAction(t *testing.T) {
	t.Run("least_outstanding_tasks picks the least loaded instance", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// POST /discover/resolve {"skill":"invoice.extract","strategy":"least_outstanding_tasks"}
		// expect: the passing instance with the fewest outstanding tasks first
	})

	t.Run("datacenter constraint limits candidates", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// POST /discover/resolve {"skill":"invoice.extract","constraints":{"datacenter":"eu-west-1"}}
		// expect: only instances in eu-west-1
	})

	t.Run("capabilities constraint requires every listed capability", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// POST /discover/resolve {"skill":"invoice.extract","constraints":{"capabilities":["streaming"]}}
		// expect: only agents whose cards declare streaming
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		t.Skip(discoverNotImplemented)

		// POST /discover/resolve with no skill
		// expect: 400 with a validation error message
	})
}
