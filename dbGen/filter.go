package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
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

		// enforce that overloaded routine has to have mapping
		if routine.HasOverload && !hasCustomMappedName(&schemaConfig, &routine) {
			// todo return error
			common.Log("Overloaded function %s doesnt have mapping", routine.RoutineNameWithParams)
			return nil, fmt.Errorf("overloaded function %s.%s doesn't have mapping defined", routine.RoutineSchema, routine.RoutineNameWithParams)
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

func hasCustomMappedName(schemaConfig *SchemaConfig, routine *DbRoutine) bool {
	mappingInfo, exists := schemaConfig.Functions[routine.RoutineNameWithParams]
	if !exists {
		common.LogDebug("mapping for function %s doesnt exist", routine.RoutineNameWithParams)
		return false
	}

	if mappingInfo.MappedName == "" {
		common.LogDebug("mapping for function %s exists, but mapped name is not set", routine.RoutineNameWithParams)

		return false
	}

	return true
}
