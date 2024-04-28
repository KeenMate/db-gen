package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/private/common"
)

func PreprocessRoutines(routines *[]DbRoutine, config *Config) error {
	err := markOverloadedRoutines(routines, config)
	if err != nil {
		return err
	}

	return nil
}

func markOverloadedRoutines(routines *[]DbRoutine, config *Config) error {
	schemaMap := getSchemaConfigMap(config)

	// first, we find all the overloaded functions
	namesCounter := make(map[string]int)
	for _, routine := range *routines {
		// carefull about schemas
		routineKey := routine.RoutineSchema + "." + routine.RoutineName
		count, exists := namesCounter[routineKey]

		if !exists {
			namesCounter[routineKey] = 1
		}

		namesCounter[routineKey] = count + 1
	}
	overloadedFunctionCount := 0
	// select all the value, that have overload
	for i, routine := range *routines {
		// carefull about schemas
		routineKey := routine.RoutineSchema + "." + routine.RoutineName
		count, exists := namesCounter[routineKey]

		if !exists || count == 1 {
			continue
		}

		schemaConfig, exists := schemaMap[routine.RoutineSchema]

		if !exists {
			panic(fmt.Sprintf("schema config for schema %s missing", routine.RoutineSchema))
		}

		// enforce that overloaded routine has to have mapping
		if routine.HasOverload && !hasCustomMappedName(&schemaConfig, &routine) {
			// todo return error
			common.Log("Overloaded function %s doesnt have mapping", routine.RoutineNameWithParams)
			return fmt.Errorf("overloaded function %s.%s doesn't have mapping defined", routine.RoutineSchema, routine.RoutineNameWithParams)
		}

		overloadedFunctionCount++
		(*routines)[i].HasOverload = true

	}

	common.Log("Marked %d functions as overload", overloadedFunctionCount)

	return nil
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

func getTypeMappings(config *Config) map[string]mapping {
	mappings := make(map[string]mapping)

	// If there are multiple mappings to one database type, last one will be used

	for _, val := range config.Mappings {
		for _, databaseType := range val.DatabaseTypes {
			mappings[databaseType] = mapping{
				mappedFunction: val.MappingFunction,
				mappedType:     val.MappedType,
			}
		}

	}

	return mappings
}
