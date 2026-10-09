// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package conf

import "os"

const (
	EditionOSS  = "oss"
	EditionSaaS = "saas"
)

// BuiltEdition is the compile-time edition.
var BuiltEdition = EditionOSS

// Edition returns the compile-time edition, or COGNIT_EDITION when set.
func Edition() string {
	switch os.Getenv("COGNIT_EDITION") {
	case EditionSaaS:
		return EditionSaaS
	case EditionOSS:
		return EditionOSS
	}

	if BuiltEdition == EditionSaaS {
		return EditionSaaS
	}

	return EditionOSS
}

// IsSaaS reports whether this process is running the hosted SaaS edition.
func IsSaaS() bool {
	return Edition() == EditionSaaS
}
