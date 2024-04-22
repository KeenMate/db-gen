package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	dbGen "github.com/keenmate/db-gen/dbGen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var getRoutinesCmd = &cobra.Command{
	Use:   "routines [out]",
	Short: "Get routines",
	Long:  "Get routines from database and save them to file to generate later",
	Run: func(cmd *cobra.Command, args []string) {
		common.BindBoolFlag(cmd, keyDebug)
		common.BindBoolFlag(cmd, keyUseRoutinesFile)
		common.BindStringFlag(cmd, keyConnectionString)
		common.BindStringFlag(cmd, keyConfig)

		configLocation := viper.GetString("config")

		_, err := dbGen.ReadConfig(configLocation)
		if err != nil {
			common.Exit("configuration error: %s", err)
		}

		log.Printf("arguments: %s", args)

		viper.AutomaticEnv() // read in environment variables that match

		err = doGetRoutines()

		if err != nil {
			common.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(getRoutinesCmd)

	// set cli flags
	common.DefineBoolFlag(getRoutinesCmd, keyDebug, "d", false, "Print debug logs and create debug files")
	common.DefineBoolFlag(getRoutinesCmd, keyUseRoutinesFile, "", false, "Use routines file to generate code")
	common.DefineStringFlag(getRoutinesCmd, keyConnectionString, "s", "", "Connection string used to connect to database")
	common.DefineStringFlag(getRoutinesCmd, keyConfig, "c", "", "Path to configuration file")
}

func doGetRoutines() error {
	log.Printf("Getting configurations...")

	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	common.LogDebug("Debug logging is enabled")

	// because we use shared config, we need to set this
	config.UseRoutinesFile = false
	log.Printf("Getting routines...")
	routines, err := dbGen.GetRoutines(config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}

	log.Printf("Saving %d routines...", len(routines))

	// TODO show what routines changed

	err = common.SaveAsJson(config.RoutinesFile, routines)
	if err != nil {
		return fmt.Errorf("error saving routines: %s", err)
	}
	log.Printf("File saved at %s", config.RoutinesFile)

	return nil
}
