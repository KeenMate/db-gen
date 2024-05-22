package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/private/dbGen"
	"github.com/keenmate/db-gen/private/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var databaseChangesFlags = []helpers.FlagArgument{
	helpers.NewBoolFlag(keyUseRoutinesFile, "", false, "Use routines file to databaseChanges code"),
}

var databaseChangesCmd = &cobra.Command{
	Use:   "database-changes",
	Short: "show changes made to database from last generation",
	Long: `
prints differences between routines loaded from database/routines file
and routines used during generation
	`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.BindFlags(cmd, append(commonFlags, databaseChangesFlags...))
		_, err := dbGen.ReadConfig(viper.GetString(keyConfig))
		if err != nil {
			helpers.Exit("configuration error: %s", err)
		}

		viper.AutomaticEnv() // read in environment variables that match

		err = doDatabaseChanges()
		if err != nil {
			helpers.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(databaseChangesCmd)

	helpers.DefineFlags(databaseChangesCmd, append(commonFlags, databaseChangesFlags...))
}

func doDatabaseChanges() error {
	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	helpers.LogDebug("Debug logging is enabled")

	// TODO it will be ideal to load build information before loading and validating config
	buildInfo, infoExist := dbGen.LoadGenerationInformation(config)

	if !infoExist {
		return fmt.Errorf("no generation information found")
	}

	if !buildInfo.CheckVersion() {
		return nil
	}

	var routines []dbGen.DbRoutine

	log.Printf("Getting routines...")
	routines, err = dbGen.GetRoutines(config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	databaseChanges := buildInfo.GetRoutinesChanges(routines)
	log.Printf("Database changes:\n" + databaseChanges)

	return nil
}
