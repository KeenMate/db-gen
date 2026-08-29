package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/private/version"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "db-gen",
	Short: "Code generator for stored procedures and functions",
	Long: `db-gen by Keenmate s.r.o.
---------
Generates typed database-access code from PostgreSQL stored functions and
procedures using Go templates you control. Language-agnostic and built for
offline, repeatable code generation.

For more information, see github.com/keenmate/db-gen
`,
	// Handle the global --llm flag here so "db-gen --llm" prints the LLM
	// reference; otherwise print version info followed by the usual help output.
	Run: func(cmd *cobra.Command, args []string) {
		if llm, _ := cmd.Flags().GetBool("llm"); llm {
			fmt.Println(formatLlmOutput())
			return
		}
		version.PrintVersion()
		_ = cmd.Help()
	},
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

	rootCmd.PersistentFlags().Bool("llm", false, "Print a full CLI reference for LLM/AI assistants")
}

func initConfig() {

}
