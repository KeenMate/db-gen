package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	dbGen "github.com/keenmate/db-gen/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var getRoutinesCmd = &cobra.Command{
	Use:   "routines [out]",
	Short: "Get routines",
	Long:  "Get routines from database to generate later",
	Run: func(cmd *cobra.Command, args []string) {
		common.BindBoolFlag(cmd, keyDebug)
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

	log.Printf("Connecting to database...")
	conn, err := dbGen.Connect(config.ConnectionString)
	if err != nil {
		return fmt.Errorf("error connecting to database: %s", err)
	}

	log.Printf("Getting routines...")
	routines, err := dbGen.GetRoutines(conn, config)
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
