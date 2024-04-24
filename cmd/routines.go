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
		common.BindFlags(cmd, commonFlags)

		configLocation := viper.GetString("config")

		_, err := dbGen.ReadConfig(configLocation)
		if err != nil {
			common.Exit("configuration error: %s", err)
		}

		log.Printf("arguments: %s", args)

		viper.AutomaticEnv() // read in environment variables that match

		if args[0] != "" {
			viper.Set("routinesFile", args[0])
		}

		err = doGetRoutines()

		if err != nil {
			common.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(getRoutinesCmd)

	common.DefineFlags(getRoutinesCmd, commonFlags)
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
