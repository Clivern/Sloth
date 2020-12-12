// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package llm

const (
	falconModelType7B   = 32
	falconModelType40B  = 60
	falconModelType180B = 80
)

func falconModelType(numLayer uint32) string {
	switch numLayer {
	case 32:
		return "7B"
	case 60:
		return "40B"
	case 80:
		return "180B"
	default:
		return "unknown"
	}
}
