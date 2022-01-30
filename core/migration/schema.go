// Copyright 2022 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package migration provides database migrations
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

	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
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

	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	ServiceID  string `json:"serviceId"`
	Command    string `json:"command"`
	Parameters string `json:"parameters"`
	Interval   int64  `json:"interval"`
	Status     string `json:"status"`
}

// LoadFromJSON update object from json
func (c *Check) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(c, data)
}

// ConvertToJSON convert object to json
func (c *Check) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(c)
}

// Silence struct
type Silence struct {
	gorm.Model

	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	CheckID   int64     `json:"checkId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

// LoadFromJSON update object from json
func (s *Silence) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(s, data)
}

// ConvertToJSON convert object to json
func (s *Silence) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(s)
}

// Metric struct
type Metric struct {
	gorm.Model

	UUID    string  `json:"uuid"`
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Tags    string  `json:"tags"`
	CheckID int64   `json:"checkId"`
}

// LoadFromJSON update object from json
func (m *Metric) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(m, data)
}

// ConvertToJSON convert object to json
func (m *Metric) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(m)
}

// Alert struct
type Alert struct {
	gorm.Model

	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	CheckID int64  `json:"checkId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// LoadFromJSON update object from json
func (a *Alert) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(a, data)
}

// ConvertToJSON convert object to json
func (a *Alert) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(a)
}

// Incident struct
type Incident struct {
	gorm.Model

	UUID        string `json:"uuid"`
	Subject     string `json:"subject"`
	AlertID     int64  `json:"alertId"`
	Description string `json:"description"`
}

// LoadFromJSON update object from json
func (i *Incident) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(i, data)
}

// ConvertToJSON convert object to json
func (i *Incident) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(i)
}

// IncidentUpdate struct
type IncidentUpdate struct {
	gorm.Model

	UUID        string `json:"uuid"`
	Subject     string `json:"subject"`
	IncidentID  int64  `json:"incidentId"`
	Description string `json:"description"`
}

// LoadFromJSON update object from json
func (u *IncidentUpdate) LoadFromJSON(data []byte) error {
	return utils.LoadFromJSON(u, data)
}

// ConvertToJSON convert object to json
func (u *IncidentUpdate) ConvertToJSON() (string, error) {
	return utils.ConvertToJSON(u)
}
