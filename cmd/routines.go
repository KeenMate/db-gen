package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/private/dbGen"
	"github.com/keenmate/db-gen/private/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var getRoutinesCmd = &cobra.Command{
	Use:   "routines [out]",
	Short: "Get routines",
	Long:  "Get routines from database and save them to file to generate later",
	Run: func(cmd *cobra.Command, args []string) {
		helpers.BindFlags(cmd, commonFlags)

		configLocation := viper.GetString("config")

		_, err := dbGen.ReadConfig(configLocation)
		if err != nil {
			helpers.Exit("configuration error: %s", err)
		}

		log.Printf("arguments: %s", args)

		viper.AutomaticEnv() // read in environment variables that match

		if len(args) > 0 && args[0] != "" {
			viper.Set("routinesFile", args[0])
		}

		err = doGetRoutines()

		if err != nil {
			helpers.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(getRoutinesCmd)

	helpers.DefineFlags(getRoutinesCmd, commonFlags)
}

func doGetRoutines() error {
	log.Printf("Getting configurations...")

	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	helpers.LogDebug("Debug logging is enabled")

	// because we use shared config, we need to set this to force loading from database
	config.UseRoutinesFile = false

	log.Printf("Getting routines...")
	routines, err := dbGen.GetRoutines(config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}

	log.Printf("Saving %d routines...", len(routines))

	// TODO show what routines changed
	// TODO remove specific name

	err = dbGen.SaveRoutinesFile(routines, config)
	if err != nil {
		return fmt.Errorf("error saving routines: %s", err)
	}
	log.Printf("File saved at %s", config.RoutinesFile)

	return nil
}
