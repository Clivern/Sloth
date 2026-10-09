// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/cognit/db"
)

// Cache is a key-value-backed cache.
type Cache struct {
	store db.KeyValueRepository
}

// NewCache returns a key-value-backed cache.
func NewCache(store db.KeyValueRepository) *Cache {
	return &Cache{store: store}
}

// Get returns a value for key, or empty if missing/expired. TODO: remove if not used in future.
func (c *Cache) Get(key string) (string, *time.Time, error) {
	item, err := c.store.Get(key)
	if err != nil {
		return "", nil, err
	}
	if item == nil {
		return "", nil, nil
	}

	return item.Value, item.ExpiresAt, nil
}

// Set stores a value for key. TODO: remove if not used in future.
func (c *Cache) Set(key, value string, expiresAt *time.Time) error {
	return c.store.Upsert(&db.KeyValue{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
	})
}

// DeleteExpired removes expired cache entries. TODO: remove if not used in future.
func (c *Cache) DeleteExpired() (int64, error) {
	return c.store.DeleteExpired()
}
