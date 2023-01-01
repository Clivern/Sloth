// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package service

import (
	"testing"

	"github.com/franela/goblin"
)

// TestUnitCorrelation
func TestUnitCorrelation(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("#Correlation", func() {
		g.It("It should satisfy test cases", func() {
			c := NewCorrelation()
			uuid := c.UUIDv4()

			g.Assert(uuid != "").Equal(true)

			for i := 0; i < 1000; i++ {
				g.Assert(c.UUIDv4() == uuid).Equal(false)
			}
		})
	})
}
