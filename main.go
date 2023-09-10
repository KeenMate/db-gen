package main

import (
	_ "database/sql"
	"db-gen/src"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

func main() {
	dbGen.Hello()
	log.Println("Starting...")

	log.Printf("Getting configurations...")
	config, err := dbGen.GetConfig()
	if err != nil {
		dbGen.Exit("error getting config %s", err)
	}
	dbGen.VerboseLog("Verbose logging is enabled")

	switch config.Command {
	case dbGen.Gen:
		doGenerate(config)
	case dbGen.Init:
		doInit(config)
	default:
		dbGen.Exit("Unknown command %s", config.Command)
	}

	log.Printf("Done")
}

func doGenerate(config *dbGen.Config) {

	log.Printf("Connecting to database...")
	conn, err := dbGen.Connect(config.ConnectionString)
	if err != nil {
		log.Panicf("error connecting to database: %s", err)
	}

	log.Printf("Getting routines...")
	routines, err := dbGen.GetRoutines(conn, config)
	if err != nil {
		dbGen.Exit("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	dbGen.VerboseLog("Saving to debug file...")
	err = dbGen.SaveToTempFile(routines, "dbRoutines")
	if err != nil {
		dbGen.Exit("error saving debug file: %s", err)
	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Preprocess(routines, config)
	if err != nil {
		dbGen.Exit("error preprocessing: %s", err)
	}
	log.Printf("After preprocessing %d functions left", len(processedFunctions))

	dbGen.VerboseLog("Saving to debug file...")
	err = dbGen.SaveToTempFile(processedFunctions, "mapped")
	if err != nil {
		dbGen.Exit("error saving debug file: %s", err)
	}

	log.Printf("Generating...")
	err = dbGen.Generate(processedFunctions, config)
	if err != nil {
		dbGen.Exit("Error generating: %s", err)
	}
}

func doInit(config *dbGen.Config) {
	log.Printf("Not implemented yet")
}
