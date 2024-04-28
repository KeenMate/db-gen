package dbGen

import (
	"github.com/keenmate/db-gen/private/common"
)

func FilterFunctions(routines *[]DbRoutine, config *Config) ([]DbRoutine, error) {
	schemaMap := getSchemaConfigMap(config)
	common.LogDebug("Got %d schema configs  ", len(schemaMap))
	filteredFunctions := make([]DbRoutine, 0)

	for _, routine := range *routines {
		schemaConfig, exists := schemaMap[routine.RoutineSchema]

		// if config for given schema doest exits, don't generate for any routine in given scheme
		if !exists {
			common.LogDebug("No schema config for '%s'", routine.RoutineSchema)
			continue
		}

		if !functionShouldBeGenerated(routine.RoutineName, &schemaConfig) {
			continue
		}

		filteredFunctions = append(filteredFunctions, routine)

	}

	return filteredFunctions, nil
}

func functionShouldBeGenerated(functionName string, schemaConfig *SchemaConfig) bool {
	// set explicitly
	val, contains := schemaConfig.Functions[functionName]
	if !contains {
		return schemaConfig.AllFunctions
	}

	common.LogDebug("Function %s has generation explicitly set to %t", functionName, val.Generate)
	return val.Generate
}

func getSchemaConfigMap(config *Config) map[string]SchemaConfig {
	schemaMap := make(map[string]SchemaConfig)

	for _, schemaConfig := range config.Generate {
		schemaMap[schemaConfig.Schema] = schemaConfig
	}

	return schemaMap
}
