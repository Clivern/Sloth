// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package service provides system wide services
package service

import (
	"github.com/DataDog/gohai/cpu"
	"github.com/DataDog/gohai/memory"
	"github.com/DataDog/gohai/platform"
)

// SystemInfo type
type SystemInfo struct {
}

// NewSystemInfo returns a new instance
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{}
}

// GetPlatformInfo ..
func (s *SystemInfo) GetPlatformInfo() (interface{}, error) {
	p := &platform.Platform{}
	return p.Collect()
}

// GetMemoryInfo ..
func (s *SystemInfo) GetMemoryInfo() (interface{}, error) {
	m := &memory.Memory{}
	return m.Collect()
}

// GetCPUInfo ..
func (s *SystemInfo) GetCPUInfo() (interface{}, error) {
	c := cpu.Cpu{}
	return c.Collect()
}
