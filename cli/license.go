// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Print the license",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Copyright (c) 2026 Clivern")
	},
}

// init registers the license subcommand.
func init() {
	rootCmd.AddCommand(licenseCmd)
}
