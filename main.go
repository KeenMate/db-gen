package main

import (
	_ "database/sql"
	"db-gen/src"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

const dbConnectString = "postgresql://postgres:Password3000!!@localhost:5432/db_gen"

func main() {
	log.Println("db_gen starting...")

	conn, err := dbGen.Connect(dbConnectString)
	if err != nil {
		log.Panic("error connecting to database")
	}

	routines, err := conn.GetRoutines("public")
	if err != nil {
		log.Panicf("error getting routines: %s", err)
	}

	log.Printf("Finished getting routines, got %d", len(routines))

	for i, routine := range routines {
		err := conn.AddParamsToRoutine(&routines[i])

		if err != nil {
			log.Panicf("error getting routine params for routine %s: %s", routine.RoutineName, err)
		}
	}

	log.Printf("Finished getting params for routines")
	err = dbGen.SaveToTempFile(routines, "dbRoutines")
	if err != nil {
		log.Printf("error savinf debug file: %s", err)
	}

	processedFunctions, err := dbGen.Preprocess(routines)

	log.Printf("Finished processing function")
	err = dbGen.SaveToTempFile(processedFunctions, "mapped")
	if err != nil {
		log.Printf("error savinf debug file: %s", err)
	}

	err = dbGen.Generate(processedFunctions)
	if err != nil {
		log.Panic(err)
	}
}
