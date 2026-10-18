// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/clivern/sloth/core"

	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the Sloth NATS worker",
	Run: func(_ *cobra.Command, _ []string) {
		err := core.Load(config)
		if err != nil {
			panic(err.Error())
		}

		err = core.SetupLogging()
		if err != nil {
			panic(err.Error())
		}

		err = core.RunWorker()
		if err != nil {
			panic(fmt.Sprintf("Worker error: %s", err.Error()))
		}
	},
}

// init registers the worker subcommand and flags.
func init() {
	workerCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"mgmt_config.prod.yml",
		"Absolute path to config file (required)",
	)
	workerCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(workerCmd)
}
