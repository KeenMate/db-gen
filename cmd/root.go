package cmd

import (
	"github.com/keenmate/db-gen/common"
	dbGen "github.com/keenmate/db-gen/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "db-gen",
	Short: "Code generator for stored procedures and functions",
	Long: `DB-GEN by KEEN|MATE
---------
TODO better description
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(versionStringFile string) {
	// because this is a top level file, we have to pass it like this
	_ = dbGen.ParseBuildInformation(versionStringFile)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//common.ConfigurationString(rootCmd, "config", "c", "", "Path to configuration file")
	//common.ConfigurationBool(rootCmd, "debug", "d", false, "Print debug information")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configLocation := viper.GetString("config")
	settings := viper.AllSettings()
	common.Log("%+v", settings)
	_, err := dbGen.ReadConfig(configLocation)
	if err != nil {
		dbGen.Exit("configuration error: %s", err)
	}

	viper.AutomaticEnv() // read in environment variables that match
}
