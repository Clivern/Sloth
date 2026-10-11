// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValidationStruct struct {
	Email    string `json:"email"    validate:"required,email"`
	URL      string `json:"url"      validate:"required,url"`
	Password string `json:"password" validate:"required,strong_password"`
	Name     string `json:"name"     validate:"required,min=2,max=50"`
	Age      int    `json:"age"      validate:"omitempty,gte=0,lte=150"`
}

type jsonFieldStruct struct {
	JSON       string `json:"json"        validate:"json"`
	JSONObject string `json:"json_object" validate:"json_object"`
}

func TestUnitValidation(t *testing.T) {
	t.Run("ValidateStruct", func(t *testing.T) {
		cases := []struct {
			name      string
			data      testValidationStruct
			wantErr   bool
			errFields []string
		}{
			{
				name: "valid struct",
				data: testValidationStruct{
					Email:    "test@example.com",
					URL:      "https://example.com",
					Password: "SecurePass123!",
					Name:     "John Doe",
					Age:      30,
				},
			},
			{
				name: "invalid email",
				data: testValidationStruct{
					Email:    "invalid-email",
					URL:      "https://example.com",
					Password: "SecurePass123!",
					Name:     "John",
				},
				wantErr:   true,
				errFields: []string{"email"},
			},
			{
				name: "invalid URL",
				data: testValidationStruct{
					Email:    "test@example.com",
					URL:      "not-a-url",
					Password: "SecurePass123!",
					Name:     "John",
				},
				wantErr:   true,
				errFields: []string{"URL"},
			},
			{
				name: "weak password",
				data: testValidationStruct{
					Email:    "test@example.com",
					URL:      "https://example.com",
					Password: "weak",
					Name:     "John",
				},
				wantErr:   true,
				errFields: []string{"password"},
			},
			{
				name:    "missing required fields",
				data:    testValidationStruct{},
				wantErr: true,
			},
			{
				name: "name too short",
				data: testValidationStruct{
					Email:    "test@example.com",
					URL:      "https://example.com",
					Password: "SecurePass123!",
					Name:     "J",
				},
				wantErr:   true,
				errFields: []string{"name", "at least 2"},
			},
			{
				name: "age out of range",
				data: testValidationStruct{
					Email:    "test@example.com",
					URL:      "https://example.com",
					Password: "SecurePass123!",
					Name:     "John",
					Age:      200,
				},
				wantErr:   true,
				errFields: []string{"age"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateStruct(tc.data)
				if tc.wantErr {
					assert.Error(t, err)
					msg := FormatValidationErrors(err)
					assert.NotEmpty(t, msg)
					for _, field := range tc.errFields {
						assert.Contains(t, msg, field)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Strong password validator", func(t *testing.T) {
		type passwordTest struct {
			Password string `validate:"strong_password"`
		}

		cases := []struct {
			name     string
			password string
			valid    bool
		}{
			{name: "valid with all requirements", password: "Password123!", valid: true},
			{name: "valid with multiple special chars", password: "P@ssw0rd!#$", valid: true},
			{name: "valid exactly 8 characters", password: "Pass123!", valid: true},
			{name: "missing uppercase", password: "password123!", valid: false},
			{name: "missing lowercase", password: "PASSWORD123!", valid: false},
			{name: "missing digit", password: "Password!@#", valid: false},
			{name: "missing special character", password: "Password123", valid: false},
			{name: "too short", password: "Pass12!", valid: false},
			{name: "empty string", password: "", valid: false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateStruct(passwordTest{Password: tc.password})
				if tc.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("JSON field validators", func(t *testing.T) {
		cases := []struct {
			name  string
			data  jsonFieldStruct
			valid bool
		}{
			{
				name:  "valid json and object",
				data:  jsonFieldStruct{JSON: `["a"]`, JSONObject: `{"k":"v"}`},
				valid: true,
			},
			{name: "empty strings allowed", data: jsonFieldStruct{}, valid: true},
			{name: "invalid json", data: jsonFieldStruct{JSON: `{bad`}, valid: false},
			{
				name:  "json array not object",
				data:  jsonFieldStruct{JSONObject: `["a"]`},
				valid: false,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateStruct(tc.data)
				if tc.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

type testRequestBody struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name"  validate:"required,min=2,max=50"`
}

func TestUnitHTTPValidation(t *testing.T) {
	t.Run("DecodeAndValidate", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/",
			strings.NewReader(`{"email":"test@example.com","name":"John"}`),
		)
		var data testRequestBody
		assert.NoError(t, DecodeAndValidate(req, &data))
		assert.Equal(t, "test@example.com", data.Email)

		invalidJSON := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`))
		var badJSON testRequestBody
		assert.Error(t, DecodeJSON(invalidJSON, &badJSON))

		invalidData := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"bad","name":"J"}`))
		var invalid testRequestBody
		assert.Error(t, DecodeAndValidate(invalidData, &invalid))
	})

	t.Run("WriteValidationError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"bad","name":"J"}`))
		var invalid testRequestBody
		err := DecodeAndValidate(req, &invalid)

		w := httptest.NewRecorder()
		WriteValidationError(w, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp["errorMessage"])

		wGeneric := httptest.NewRecorder()
		WriteValidationError(wGeneric, bytes.ErrTooLarge)
		assert.Equal(t, http.StatusBadRequest, wGeneric.Code)
		assert.Contains(t, wGeneric.Body.String(), bytes.ErrTooLarge.Error())
	})
}
