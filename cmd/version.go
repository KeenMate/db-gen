package cmd

import (
	_ "embed"
	"github.com/keenmate/db-gen/private/version"
	"github.com/spf13/cobra"
)

// fromDatabaseCommand represents the toDatabase command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints versions",
	Long:  `Prints executable version and build information`,
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
