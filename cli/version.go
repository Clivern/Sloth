// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package cli

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version buildinfo item
	Version = "dev"
	// Commit buildinfo item
	Commit = "none"
	// Date buildinfo item
	Date = "unknown"
	// BuiltBy buildinfo item
	BuiltBy = "unknown"
	// Edition buildinfo item
	Edition = "oss"
	// Static embedded files
	Static embed.FS
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf(
			"Current Sloth version %v commit %v, built @%v by %v, edition %v.\n",
			Version,
			Commit,
			Date,
			BuiltBy,
			Edition,
		)
	},
}

// init registers the version subcommand.
func init() {
	rootCmd.AddCommand(versionCmd)
}
