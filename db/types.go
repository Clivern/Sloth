// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// Id is a PostgreSQL UUID primary or foreign key.
// It implements sql.Scanner and driver.Valuer so UUIDs can be scanned into DB structs.
type Id string

// NewId returns a new UUID-backed database Id.
func NewId() (Id, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}

	return Id(id.String()), nil
}

// Scan implements sql.Scanner. Accepts []byte or string UUID values.
func (id *Id) Scan(value interface{}) error {
	if value == nil {
		*id = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*id = Id(string(v))
		return nil
	case string:
		*id = Id(v)
		return nil
	default:
		return fmt.Errorf("db: cannot scan type %T into Id", value)
	}
}

// Value implements driver.Valuer for INSERT/UPDATE.
func (id Id) Value() (driver.Value, error) {
	if lo.IsEmpty(id) {
		return nil, nil
	}

	return string(id), nil
}

// String returns the Id as string.
func (id Id) String() string {
	return string(id)
}

// IsBotUser reports whether id is the platform bot user.
func IsBotUser(id Id) bool {
	return id == BotUserId
}
