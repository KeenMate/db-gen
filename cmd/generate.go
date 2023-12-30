package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	dbGen "github.com/keenmate/db-gen/src"
	"github.com/spf13/cobra"
	"log"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code",
	Long:  "Generate code for calling database stored procedures",
	Run: func(cmd *cobra.Command, args []string) {
		err := doGenerate()
		if err != nil {
			dbGen.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// set cli flags
	common.ConfigurationBool(generateCmd, "debug", "d", false, "Print debug logs and create debug files")
	common.ConfigurationString(generateCmd, "connectionString", "s", "", "Connection string used to connect to database")
	common.ConfigurationString(generateCmd, "config", "c", "", "Path to configuration file")

}

func doGenerate() error {
	log.Printf("Getting configurations...")

	config, err := dbGen.GetConfig()
	if err != nil {
		return fmt.Errorf("error getting config %s", err)
	}

	common.LogDebug("Verbose logging is enabled")

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
	log.Printf("Got %d routines", len(routines))

	if config.Verbose {
		common.LogDebug("Saving to debug file...")
		err = dbGen.SaveToTempFile(routines, "dbRoutines")
		if err != nil {
			return fmt.Errorf("error saving debug file: %s", err)
		}
	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Preprocess(routines, config)
	if err != nil {
		return fmt.Errorf("error preprocessing: %s", err)
	}
	log.Printf("After preprocessing %d - %d = %d functions left", len(routines), len(routines)-len(processedFunctions), len(processedFunctions))

	if config.Verbose {
		common.LogDebug("Saving to debug file...")
		err = dbGen.SaveToTempFile(processedFunctions, "mapped")
		if err != nil {
			return fmt.Errorf("error saving debug file: %s", err)
		}
	}

	log.Printf("Generating...")
	err = dbGen.Generate(processedFunctions, config)
	if err != nil {
		return fmt.Errorf("error generating: %s", err)
	}

	return nil
}
