// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package resend

import (
	"bytes"
	"fmt"
	"html/template"
)

// RenderTemplate renders an email template body with data.
func RenderTemplate(body string, data any) (string, error) {
	tmpl, err := template.New("mail").Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse mail template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("render mail template: %w", err)
	}

	return buf.String(), nil
}
