// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

var (
	// rwConn holds the primary read-write database connection.
	rwConn *Connection
	// roConns holds read-only database connections used for load-balanced reads.
	roConns []*Connection
	// nxtConn tracks the next connection for round-robin read balancing.
	nxtConn atomic.Uint64
	// mu protects database connections during initialization and shutdown.
	mu sync.RWMutex
)

// InitDB initializes the global database connections.
func InitDB(rwConfig DatabaseConfig, roConfigs ...DatabaseConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if rwConn != nil {
		log.Warn().
			Msg("Database connection already initialized")
		return nil
	}

	conn, err := NewConnection(rwConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize read-write database: %w", err)
	}

	iroConn := make([]*Connection, 0, len(roConfigs))
	for index, config := range roConfigs {
		roConn, err := NewConnection(config)
		if err != nil {
			conn.Close()
			CloseConnections(iroConn)
			return fmt.Errorf("failed to initialize read-only database %d: %w", index, err)
		}

		iroConn = append(iroConn, roConn)
	}

	rwConn = conn
	roConns = iroConn
	nxtConn.Store(0)

	log.Info().
		Int("read_only_connections", len(roConns)).
		Msg("Global database connections initialized")

	return nil
}

// GetDB returns a database connection.
// By default it load-balances reads across read-write and read-only connections.
// Pass true to enforce consistency and use the read-write connection.
func GetDB(enforceConsistency ...bool) *sql.DB {
	if len(enforceConsistency) > 0 && enforceConsistency[0] {
		return GetWriteDB()
	}

	return GetReadDB()
}

// GetWriteDB returns the read-write database connection.
func GetWriteDB() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()

	return rwConn.DB
}

// GetReadDB round-robins across the read-write connection and configured read-only connections.
func GetReadDB() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()

	if len(roConns) == 0 {
		return rwConn.DB
	}

	index := nxtConn.Add(1) % uint64(len(roConns)+1)
	if index == 0 {
		return rwConn.DB
	}

	return roConns[index-1].DB
}

// CloseDB closes the global database connections.
func CloseDB() error {
	mu.Lock()
	defer mu.Unlock()

	if rwConn == nil {
		return nil
	}

	err := rwConn.Close()
	if roErr := CloseConnections(roConns); roErr != nil && err == nil {
		err = roErr
	}

	rwConn = nil
	roConns = nil
	nxtConn.Store(0)

	return err
}

// CloseConnections closes the database connections.
func CloseConnections(connections []*Connection) error {
	var closeErr error

	for _, conn := range connections {
		err := conn.Close()
		if err != nil && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}
