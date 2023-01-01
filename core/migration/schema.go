// Copyright 2022 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package migration

import (
	"time"

	"github.com/clivern/sloth/core/utils"

	"github.com/jinzhu/gorm"
)

// Option struct
type Option struct {
	gorm.Model

	Name  string `json:"name"`
	Value string `json:"value"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoadFromJSON update object from json
func (o *Option) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(o, data)
}

// ConvertToJSON convert object to json
func (o *Option) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(o)
}

// User struct
type User struct {
	gorm.Model

	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoadFromJSON update object from json
func (u *User) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(u, data)
}

// ConvertToJSON convert object to json
func (u *User) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(u)
}

// Service struct
type Service struct {
	gorm.Model

	Name string `json:"name"`
	Slug string `json:"slug"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoadFromJSON update object from json
func (s *Service) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(s, data)
}

// ConvertToJSON convert object to json
func (s *Service) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(s)
}

// Check struct
type Check struct {
	gorm.Model

	Name string `json:"name"`
	Slug string `json:"slug"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoadFromJSON update object from json
func (c *Check) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(c, data)
}

// ConvertToJSON convert object to json
func (c *Check) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(c)
}
