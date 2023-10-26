package main

import (
	_ "database/sql"
	"db-gen/src"
	_ "embed"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

//go:embed version.txt
var versionFile string

func main() {
	err := dbGen.ParseBuildInformation(versionFile)
	if err != nil {
		dbGen.Exit("%s", err)
	}

	dbGen.Hello()
	log.Println("Starting...")
	command := dbGen.GetCommand()
	switch command {
	case dbGen.Gen:
		err = doGenerate()
		if err != nil {
			dbGen.Exit("error generating %s", err)
		}
	case dbGen.Init:
		doInit()
	case dbGen.Version:
		dbGen.PrintVersion()
	default:
		dbGen.Exit("Unknown command %s", command)
	}

	log.Printf("Done")
}

func doGenerate() error {

	log.Printf("Getting configurations...")

	config, err := dbGen.GetConfig()
	if err != nil {
		dbGen.Exit("error getting config %s", err)
	}

	dbGen.VerboseLog("Verbose logging is enabled")

	log.Printf("Connecting to database...")
	conn, err := dbGen.Connect(config.ConnectionString)
	if err != nil {
		log.Panicf("error connecting to database: %s", err)
	}

	log.Printf("Getting routines...")
	routines, err := dbGen.GetRoutines(conn, config)
	if err != nil {
		return fmt.Errorf("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	dbGen.VerboseLog("Saving to debug file...")
	err = dbGen.SaveToTempFile(routines, "dbRoutines")
	if err != nil {
		return fmt.Errorf("error saving debug file: %s", err)
	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Preprocess(routines, config)
	if err != nil {
		return fmt.Errorf("error preprocessing: %s", err)
	}
	log.Printf("After preprocessing %d functions left", len(processedFunctions))

	dbGen.VerboseLog("Saving to debug file...")
	err = dbGen.SaveToTempFile(processedFunctions, "mapped")
	if err != nil {
		return fmt.Errorf("error saving debug file: %s", err)
	}

	log.Printf("Generating...")
	err = dbGen.Generate(processedFunctions, config)
	if err != nil {
		return fmt.Errorf("error generating: %s", err)
	}

	return nil
}

func doInit() {
	dbGen.Exit("Init is not implemented yet")
}
