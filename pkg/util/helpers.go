// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// AppURL returns the configured app URL joined with a relative path.
func AppURL(path string) string {
	return strings.TrimRight(viper.GetString("app.url"), "/") + path
}

// CurrentMonthPeriod returns the UTC start and end of the current calendar month.
func CurrentMonthPeriod() (time.Time, time.Time) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	return start, start.AddDate(0, 1, 0)
}

// WriteJSON writes a value as JSON to an HTTP response.
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON response: %w", err)
	}

	return nil
}

// GenerateUUID returns a new random UUID string.
func GenerateUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}

	return id.String(), nil
}

// GenerateSecureToken returns a cryptographically secure random token.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("generate secure token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// RandInt returns a random integer in [0, n).
func RandInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, fmt.Errorf("generate random int: %w", err)
	}

	return min + int(n.Int64()), nil
}

// ParseQueryLabels extracts label key-value pairs from query parameters.
// Reserved keys such as id, limit, and offset are skipped.
func ParseQueryLabels(r *http.Request) map[string]string {
	labels := map[string]string{}

	for key, values := range r.URL.Query() {
		if key == "id" || key == "limit" || key == "offset" {
			continue
		}
		if len(values) == 0 {
			continue
		}

		labels[key] = values[0]
	}

	return labels
}

// ParsePagination parses limit and offset from query parameters.
func ParsePagination(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if v := r.URL.Query().Get("limit"); lo.IsNotEmpty(v) {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := r.URL.Query().Get("offset"); lo.IsNotEmpty(v) {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// RandomHandle returns a handle like abcd-1234 (min–max letters, dash, min–max digits).
func RandomHandle(min, max int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	const digits = "0123456789"

	letterLen, err := RandInt(min, max)
	if err != nil {
		return "", err
	}

	digitLen, err := RandInt(min, max)
	if err != nil {
		return "", err
	}

	handle := make([]byte, letterLen+1+digitLen)
	for i := 0; i < letterLen; i++ {
		idx, err := RandInt(0, len(letters)-1)
		if err != nil {
			return "", err
		}

		handle[i] = letters[idx]
	}

	handle[letterLen] = '-'
	for i := 0; i < digitLen; i++ {
		idx, err := RandInt(0, len(digits)-1)
		if err != nil {
			return "", err
		}

		handle[letterLen+1+i] = digits[idx]
	}

	return string(handle), nil
}

// HandleFromName derives a dash-separated lowercase handle from a display name.
func HandleFromName(name string, maxLength int) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if b.Len() > 0 && !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	handle := strings.Trim(b.String(), "-")
	if maxLength > 0 && len(handle) > maxLength {
		handle = strings.TrimRight(handle[:maxLength], "-")
	}

	return handle
}

// RemoveLabelFromJSON removes a label from a JSON-encoded string array.
func RemoveLabelFromJSON(labels *string, label string) (*string, bool) {
	if labels == nil || lo.IsEmpty(*labels) {
		return labels, false
	}

	var items []string
	if err := json.Unmarshal([]byte(*labels), &items); err != nil {
		return labels, false
	}
	if !lo.Contains(items, label) {
		return labels, false
	}

	next := lo.Without(items, label)
	if len(next) == 0 {
		return nil, true
	}

	raw, err := json.Marshal(next)
	if err != nil {
		return labels, false
	}

	s := string(raw)

	return &s, true
}

// JSONRawFromString returns stored JSON text as a RawMessage for API responses.
func JSONRawFromString(raw *string) json.RawMessage {
	if raw == nil || lo.IsEmpty(*raw) {
		return nil
	}

	return json.RawMessage(*raw)
}

// JSONSliceFromString unmarshals a JSON-encoded slice from a stored string pointer.
func JSONSliceFromString[T any](raw *string) []T {
	if raw == nil || lo.IsEmpty(*raw) {
		return nil
	}

	var items []T
	if err := json.Unmarshal([]byte(*raw), &items); err != nil {
		return nil
	}

	return items
}
