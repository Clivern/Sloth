// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package ai

import (
	"errors"
)

var (
	ErrEmbedNotConfigured = errors.New("ai embeddings are not configured")
	ErrLLMNotConfigured   = errors.New("ai llm is not configured")
	ErrNoEmbeddings       = errors.New("ai returned no embeddings")
	ErrNoCompletion       = errors.New("ai returned no completion")
)
