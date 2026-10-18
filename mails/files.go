// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package mails

import _ "embed"

//go:embed invite.html
var Invite string

//go:embed welcome.html
var Welcome string

//go:embed verify-email.html
var VerifyEmail string

//go:embed reset-pwd.html
var ResetPwd string
