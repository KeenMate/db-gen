package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	"log"
)

// Transforms data returned by database to structures that are used in generator

func Process(routines []DbRoutine, config *Config) ([]Routine, error) {

	filteredRoutines, err := FilterFunctions(&routines, config)
	if err != nil {
		return nil, fmt.Errorf("filtering routines: %s", err)
	}

	err = PreprocessRoutines(&routines, config)

	// don't need to compute for every property
	typeMappings := getTypeMappings(config)
	common.LogDebug("Got %d type mappings", len(typeMappings))

	// Map routines
	functions, err := mapFunctions(&filteredRoutines, &typeMappings, config)

	if err != nil {
		log.Println("Error while processing functions")
		return nil, fmt.Errorf("mapping functions: %s", err)
	}

	return functions, nil

}
