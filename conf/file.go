// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package conf

import (
	"github.com/clivern/sloth/mails"
)

var (
	// InviteEmail is the workspace invite email body.
	InviteEmail = mails.Invite

	// WelcomeEmail is the account welcome email body.
	WelcomeEmail = mails.Welcome

	// VerifyEmail is the email verification body.
	VerifyEmail = mails.VerifyEmail

	// ResetPwdEmail is the password reset email body.
	ResetPwdEmail = mails.ResetPwd
)
