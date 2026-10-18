// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
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
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to get cache key")
		return "", nil, err
	}
	if item == nil {
		return "", nil, nil
	}

	return item.Value, item.ExpiresAt, nil
}

// Set stores a value for key. TODO: remove if not used in future.
func (c *Cache) Set(key, value string, expiresAt *time.Time) error {
	err := c.store.Upsert(&db.KeyValue{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to set cache key")
		return err
	}

	log.Debug().
		Str("key", key).
		Msg("Cache key set")

	return nil
}

// DeleteExpired removes expired cache entries. TODO: remove if not used in future.
func (c *Cache) DeleteExpired() (int64, error) {
	count, err := c.store.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired cache keys")
		return 0, err
	}

	if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired cache keys deleted")
	}

	return count, nil
}
