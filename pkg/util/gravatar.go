// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

const gravatarBaseURL = "https://www.gravatar.com/avatar"

// GravatarURL returns the Gravatar avatar URL for the given email.
func GravatarURL(email string, size int) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if lo.IsEmpty(email) {
		return ""
	}

	sum := md5.Sum([]byte(email))
	hash := hex.EncodeToString(sum[:])
	url := fmt.Sprintf("%s/%s?d=identicon", gravatarBaseURL, hash)

	if size > 0 && size <= 2048 {
		url += fmt.Sprintf("&s=%d", size)
	}

	return url
}
