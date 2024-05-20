package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	dbGen "github.com/keenmate/db-gen/dbGen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const keyUseRoutinesFile = "useRoutinesFile"

var generateFlags = []common.FlagArgument{
	common.NewBoolFlag(keyUseRoutinesFile, "", false, "Use routines file to generate code"),
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
		common.BindFlags(cmd, append(commonFlags, generateFlags...))
		_, err := dbGen.ReadConfig(viper.GetString(keyConfig))
		if err != nil {
			common.Exit("configuration error: %s", err)
		}

		viper.AutomaticEnv() // read in environment variables that match

		err = doGenerate()
		if err != nil {
			common.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	common.DefineFlags(generateCmd, append(commonFlags, generateFlags...))
}

func doGenerate() error {
	timer := common.NewTimer()
	log.Printf("Getting configurations...")

	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	timer.AddEntry("getting config")
	common.LogDebug("Debug logging is enabled")

	var routines []dbGen.DbRoutine

	log.Printf("Getting routines...")
	routines, err = dbGen.GetRoutines(config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	timer.AddEntry("getting routines")
	if config.Debug {
		common.LogDebug("Saving to debug file...")
		err = common.SaveToTempFile(routines, "dbRoutines")
		if err != nil {
			return fmt.Errorf("error saving debug file: %s", err)
		}
		timer.AddEntry("saving debug file")

	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Process(routines, config)
	if err != nil {
		return fmt.Errorf("error preprocessing: %s", err)
	}
	log.Printf("After preprocessing %d - %d = %d functions left", len(routines), len(routines)-len(processedFunctions), len(processedFunctions))
	timer.AddEntry("preprocessing")

	if config.Debug {
		common.LogDebug("Saving to debug file...")
		err = common.SaveToTempFile(processedFunctions, "mapped")
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
	timer.Finish()
	log.Printf(timer.String())
	return nil
}
