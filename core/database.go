// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"github.com/clivern/sloth/db"

	"github.com/spf13/viper"
)

// Database holds the database configuration
type Database struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"name"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// ReadWriteDatabase returns the primary read-write database configuration.
func ReadWriteDatabase() db.DatabaseConfig {
	return databaseConfig("app.database.rw")
}

// ReadOnlyDatabase returns read-only database configurations.
func ReadOnlyDatabase() []db.DatabaseConfig {
	var databases []Database
	ro := []db.DatabaseConfig{}

	_ = viper.UnmarshalKey("app.database.ro", &databases)

	for _, database := range databases {
		ro = append(ro, database.DBConfig())
	}

	return ro
}

// DBConfig converts a core database config to a db package config.
func (d Database) DBConfig() db.DatabaseConfig {
	return db.DatabaseConfig{
		Driver:          d.Driver,
		Host:            d.Host,
		Port:            d.Port,
		Username:        d.Username,
		Password:        d.Password,
		Database:        d.Database,
		MaxOpenConns:    d.MaxOpenConns,
		MaxIdleConns:    d.MaxIdleConns,
		ConnMaxLifetime: d.ConnMaxLifetime,
	}
}

// databaseConfig loads database settings from configuration.
func databaseConfig(prefix string) db.DatabaseConfig {
	return db.DatabaseConfig{
		Driver:          viper.GetString(prefix + ".driver"),
		Host:            viper.GetString(prefix + ".host"),
		Port:            viper.GetInt(prefix + ".port"),
		Username:        viper.GetString(prefix + ".username"),
		Password:        viper.GetString(prefix + ".password"),
		Database:        viper.GetString(prefix + ".name"),
		MaxOpenConns:    viper.GetInt(prefix + ".max_open_conns"),
		MaxIdleConns:    viper.GetInt(prefix + ".max_idle_conns"),
		ConnMaxLifetime: viper.GetInt(prefix + ".conn_max_lifetime"),
	}
}
