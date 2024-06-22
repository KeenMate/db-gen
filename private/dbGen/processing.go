package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/private/helpers"
	"log"
)

// Transforms data returned by database to structures that are used in generator

func Process(routines []DbRoutine, config *Config) ([]Routine, error) {

	filteredRoutines, err := FilterFunctions(&routines, config)
	if err != nil {
		return nil, fmt.Errorf("filtering routines: %s", err)
	}

	// don't need to compute for every property
	typeMappings := getTypeMappings(config)
	helpers.LogDebug("Got %d type mappings", len(typeMappings))

	// Map routines
	functions, err := mapRoutines(&filteredRoutines, &typeMappings, config)

	if err != nil {
		log.Println("Error while processing functions")
		return nil, fmt.Errorf("mapping functions: %s", err)
	}

	return functions, nil

}
