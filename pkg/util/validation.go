// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// Validator is a singleton instance of the validator
var validate *validator.Validate

// init registers custom validators.
func init() {
	validate = validator.New()

	// Register custom tag name function to use 'label' tag
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		label := fld.Tag.Get("label")
		if lo.IsNotEmpty(label) {
			return label
		}

		name := fld.Tag.Get("json")
		if lo.IsNotEmpty(name) {
			return name
		}

		return fld.Name
	})

	// Register custom validators
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("json", validateJSONString)
	validate.RegisterValidation("json_object", validateJSONObjectString)
}

// GetValidator returns the shared validator instance.
func GetValidator() *validator.Validate {
	return validate
}

// ValidationError is one field error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors is a list of validation errors.
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// ValidateStruct validates the struct and returns any validation errors.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationErrors returns the first validation error message
func FormatValidationErrors(err error) string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrs) > 0 {
			return getErrorMessage(validationErrs[0])
		}
	}

	return ""
}

// getErrorMessage returns a user-friendly error message based on the validation tag
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of: %s", field, e.Param())
	case "startswith":
		return fmt.Sprintf("%s must start with %s", field, e.Param())
	case "endswith":
		return fmt.Sprintf("%s must end with %s", field, e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least 8 characters, one uppercase, one lowercase, one digit, and one special character", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "json_object":
		return fmt.Sprintf("%s must be a valid JSON object", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// validateJSONString validates that a string contains valid JSON.
func validateJSONString(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	s = strings.TrimSpace(s)
	if lo.IsEmpty(s) {
		return true
	}

	var tmp interface{}

	return json.Unmarshal([]byte(s), &tmp) == nil
}

// validateJSONObjectString validates that a string contains a JSON object.
func validateJSONObjectString(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	s = strings.TrimSpace(s)
	if lo.IsEmpty(s) {
		return true
	}

	var tmp map[string]interface{}

	return json.Unmarshal([]byte(s), &tmp) == nil
}

// validateStrongPassword validates that a password meets security requirements
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	// Has uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Has lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Has digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Has special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// DecodeJSON reads and decodes JSON from the request body.
func DecodeJSON(r *http.Request, v any) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("Invalid JSON format: %w", err)
	}

	return nil
}

// DecodeAndValidate decodes JSON and validates the struct in one step.
func DecodeAndValidate(r *http.Request, v any) error {
	err := DecodeJSON(r, v)
	if err != nil {
		return err
	}

	return ValidateStruct(v)
}

// WriteValidationError writes a decode or validation error as a 400 JSON response.
func WriteValidationError(w http.ResponseWriter, err error) {
	message := err.Error()
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		message = FormatValidationErrors(validationErrs)
	}

	WriteJSON(w, http.StatusBadRequest, map[string]any{
		"errorMessage": message,
	})
}
