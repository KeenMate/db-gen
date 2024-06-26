package cmd

import (
	"github.com/keenmate/db-gen/private/version"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "db-gen",
	Short: "Code generator for stored procedures and functions",
	Long: `DB-GEN by KEEN|MATE
---------
For more information, see github.com/keenmate/db-gen
`,
}

// Execute adds all child commands to the root command and sets generateCmdFlags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(versionStringFile string) {
	// because this is a top level file, we have to pass it like this
	_ = version.ParseBuildInformation(versionStringFile)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

}
