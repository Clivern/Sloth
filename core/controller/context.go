// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/clivern/sloth/core/module"
)

// Context type
type Context struct {
	DB *module.Database
}

// GetDatabase gets a database connection
func (g *Context) GetDatabase() *module.Database {
	return g.DB
}
