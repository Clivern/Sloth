// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package util

import "testing"

func TestHandleFromName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "simple", input: "Customer Support", expect: "customer-support"},
		{name: "symbols", input: "CI / Deploy!!!", expect: "ci-deploy"},
		{name: "empty", input: "!!!", expect: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HandleFromName(tc.input, 100); got != tc.expect {
				t.Fatalf("HandleFromName(%q) = %q, want %q", tc.input, got, tc.expect)
			}
		})
	}
}
