package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/private/dbGen"
	"github.com/keenmate/db-gen/private/helpers"
	"github.com/keenmate/db-gen/private/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const keyUseRoutinesFile = "useRoutinesFile"

var generateFlags = []helpers.FlagArgument{
	helpers.NewBoolFlag(keyUseRoutinesFile, "", false, "Use routines file to generate code"),
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code",
	Long: `
	Generate code for calling stored procedures.
	Procedures can be loaded from database or from provided file.
	Output folder and templates are defined in configuration file.
	
	For more information, see github.com/keenmate/db-gen

	`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.BindFlags(cmd, append(commonFlags, generateFlags...))
		_, err := dbGen.ReadConfig(viper.GetString(keyConfig))
		if err != nil {
			helpers.Exit("configuration error: %s", err)
		}

		viper.AutomaticEnv() // read in environment variables that match

		err = doGenerate()
		if err != nil {
			helpers.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	helpers.DefineFlags(generateCmd, append(commonFlags, generateFlags...))
}

func doGenerate() error {
	timer := helpers.NewTimer()
	log.Printf("Getting configurations...")

	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	helpers.LogDebug("Debug logging is enabled")
	timer.AddEntry("getting config")

	// TODO it will be ideal to load build information before loading and validating config
	buildInfo, infoExist := dbGen.LoadGenerationInformation(config)
	timer.AddEntry("loading build information")

	if infoExist {
		log.Printf("Build information loaded, last build was at %s", buildInfo.Time.String())

		if !buildInfo.CheckVersion() {
			return nil
		}
	}

	var routines []dbGen.DbRoutine

	log.Printf("Getting routines...")
	routines, err = dbGen.GetRoutines(config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	timer.AddEntry("getting routines")
	if config.Debug {
		helpers.LogDebug("Saving to debug file...")
		err = helpers.SaveToTempFile(routines, "dbRoutines")
		if err != nil {
			return fmt.Errorf("error saving debug file: %s", err)
		}
		timer.AddEntry("saving debug file")
	}

	if infoExist {
		// TODO we should only show changes of routines after filtering
		changesMsg := buildInfo.GetRoutinesChanges(routines)
		log.Printf("Database changes:\n" + changesMsg)
	} else {
		log.Printf("No previous build information found")
	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Process(routines, config)
	if err != nil {
		return fmt.Errorf("error preprocessing: %s", err)
	}
	log.Printf("After preprocessing %d - %d = %d functions left", len(routines), len(routines)-len(processedFunctions), len(processedFunctions))
	timer.AddEntry("preprocessing")

	if config.Debug {
		helpers.LogDebug("Saving to debug file...")
		err = helpers.SaveToTempFile(processedFunctions, "mapped")
		if err != nil {
			return fmt.Errorf("error saving debug file: %s", err)
		}
		timer.AddEntry("saving debug file")

	}

	log.Printf("Generating...")
	err = dbGen.Generate(processedFunctions, config)
	if err != nil {
		return fmt.Errorf("error generating: %s", err)
	}
	timer.AddEntry("generating files")

	err = dbGen.SaveGenerationInformation(config, routines, version.GetVersion())
	if err != nil {
		log.Printf("Error saving generation information: %v", err)
	}

	timer.AddEntry("saving generation info")
	timer.Finish()
	log.Printf(timer.String())
	return nil
}
