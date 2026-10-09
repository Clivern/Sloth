// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package cli

import (
	"github.com/clivern/cognit/core"
	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/migration"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  `Manage database migrations`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run: func(cmd *cobra.Command, _ []string) {
		configFile, _ := cmd.Flags().GetString("config")

		err := core.Load(configFile)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to load configuration")
		}

		err = core.SetupLogging()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to setup logging")
		}

		conn, err := db.NewConnection(core.ReadWriteDatabase())
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to connect to database")
		}

		defer conn.Close()

		mgr := migration.NewManager(conn.DB)

		for _, m := range migration.GetAll() {
			mgr.Register(m)
		}

		err = mgr.Up()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to run migrations")
		}

		log.Info().
			Msg("Migration completed successfully")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	Run: func(cmd *cobra.Command, _ []string) {
		configFile, _ := cmd.Flags().GetString("config")

		err := core.Load(configFile)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to load configuration")
		}

		err = core.SetupLogging()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to setup logging")
		}

		conn, err := db.NewConnection(core.ReadWriteDatabase())
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to connect to database")
		}

		defer conn.Close()

		mgr := migration.NewManager(conn.DB)

		for _, m := range migration.GetAll() {
			mgr.Register(m)
		}

		err = mgr.Down()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to roll back migration")
		}

		log.Info().
			Msg("Rollback completed successfully")
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Run: func(cmd *cobra.Command, _ []string) {
		configFile, _ := cmd.Flags().GetString("config")

		err := core.Load(configFile)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to load configuration")
		}

		err = core.SetupLogging()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to setup logging")
		}

		conn, err := db.NewConnection(core.ReadWriteDatabase())
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to connect to database")
		}

		defer conn.Close()

		mgr := migration.NewManager(conn.DB)

		for _, m := range migration.GetAll() {
			mgr.Register(m)
		}

		err = mgr.Status()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to get migration status")
		}
	},
}

// init registers the migrate subcommands and flags.
func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	migrateUpCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	migrateUpCmd.MarkFlagRequired("config")
	migrateDownCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	migrateDownCmd.MarkFlagRequired("config")
	migrateStatusCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"config.prod.yml",
		"Absolute path to config file (required)",
	)
	migrateStatusCmd.MarkFlagRequired("config")
}
