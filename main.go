package main

import (
	_ "database/sql"
	"db-gen/src"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

func main() {
	printHello()
	log.Println("Starting...")

	log.Printf("Getting configurations...")
	config, err := dbGen.GetConfig()
	if err != nil {
		log.Panic("error getting config")
	}

	if config.Verbose {
		log.Print("Verbose logging is enabled")
		dbGen.PrettyPrint(config)
	}

	switch config.Command {
	case dbGen.Gen:
		doGenerate(config)
	case dbGen.Init:
		doInit(config)
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
		dbGen.Panic("error getting routines: %s", err)
	}
	log.Printf("Got %d routines", len(routines))

	log.Printf("Saving to debug file...")
	err = dbGen.SaveToTempFile(routines, "dbRoutines")
	if err != nil {
		dbGen.Panic("error savinf debug file: %s", err)
	}

	log.Printf("Preprocessing...")
	processedFunctions, err := dbGen.Preprocess(routines, config)
	if err != nil {
		dbGen.Panic("error preprocessing: %s", err)
	}

	log.Printf("Saving to debug file...")
	err = dbGen.SaveToTempFile(processedFunctions, "mapped")
	if err != nil {
		dbGen.Panic("error savinf debug file: %s", err)
	}

	log.Printf("Generating...")
	err = dbGen.Generate(processedFunctions, config)
	if err != nil {
		dbGen.Panic("Error generating: %s", err)
	}
}

func doInit(config *dbGen.Config) {
	log.Printf("Not implemented yet")
}

func printHello() {
	figure.NewColorFigure("db-gen", "", "green", true).Print()
	fmt.Println()
}
