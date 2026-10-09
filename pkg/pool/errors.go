// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package pool

import (
	"errors"

	"github.com/panjf2000/ants/v2"
)

var (
	ErrInvalidSize = errors.New("pool: size must be greater than zero")
	ErrClosed      = ants.ErrPoolClosed
	ErrOverload    = ants.ErrPoolOverload
)
