package dbGen

import (
	"sort"
	"strconv"

	"github.com/keenmate/db-gen/private/helpers"
)

func PreprocessRoutines(routines *[]DbRoutine, config *Config) error {
	markOverloadedRoutines(routines, config)

	return nil
}

// markOverloadedRoutines finds routines that share a schema-qualified name
// (PostgreSQL overloads) and assigns each a stable, 1-based numeric suffix.
// Members are ordered by their full parameter signature so the suffix is
// deterministic across runs and database rebuilds. Routines that carry an
// explicit MappedName are still marked as overloaded (change detection relies
// on the flag) but ignore the suffix when their generated name is built.
func markOverloadedRoutines(routines *[]DbRoutine, config *Config) {
	schemaMap := getSchemaConfigMap(config)

	// group routine indices by schema-qualified name
	groups := make(map[string][]int)
	for i, routine := range *routines {
		routineKey := routine.RoutineSchema + "." + routine.RoutineName
		groups[routineKey] = append(groups[routineKey], i)
	}

	overloadedFunctionCount := 0
	for routineKey, indices := range groups {
		if len(indices) < 2 {
			continue
		}

		// deterministic order: sort overloads by their full signature so the
		// numeric suffix is stable regardless of the input order
		sort.Slice(indices, func(a, b int) bool {
			return (*routines)[indices[a]].RoutineNameWithParams < (*routines)[indices[b]].RoutineNameWithParams
		})

		anyMappedName := false
		for order, idx := range indices {
			(*routines)[idx].HasOverload = true
			(*routines)[idx].OverloadSuffix = strconv.Itoa(order + 1)
			overloadedFunctionCount++

			if hasMappedName(schemaMap, &(*routines)[idx]) {
				anyMappedName = true
			}
		}

		// A fully-unmapped overload set relies on positional numeric suffixes,
		// which can shift if a new overload with an earlier-sorting signature is
		// added. Surface that so it's visible in CI output.
		if !anyMappedName {
			helpers.LogWarn("overloaded function %s has %d overloads and no MappedName; "+
				"generated names use positional numeric suffixes (e.g. ...1, ...2) whose ordering "+
				"can shift if signatures change — set MappedName to pin names", routineKey, len(indices))
		}
	}

	helpers.Log("Marked %d functions as overload", overloadedFunctionCount)
}

// hasMappedName reports whether the routine has a non-empty MappedName
// configured (keyed by its full signature) in the schema config.
func hasMappedName(schemaMap map[string]SchemaConfig, routine *DbRoutine) bool {
	schemaConfig, ok := schemaMap[routine.RoutineSchema]
	if !ok {
		return false
	}

	mapping, ok := schemaConfig.Functions[routine.RoutineNameWithParams]
	return ok && mapping.MappedName != ""
}

func getTypeMappings(config *Config) map[string]mapping {
	mappings := make(map[string]mapping)

	// If there are multiple mappings to one database type, last one will be used

	for _, val := range config.Mappings {
		for _, databaseType := range val.DatabaseTypes {
			mappings[databaseType] = mapping{
				mappedFunction:        val.MappingFunction,
				mappedType:            val.MappedType,
				nullableReturnType:    val.NullableReturnType,
				nullableParameterType: val.NullableParameterType,
				optionalParameterType: val.OptionalParameterType,
			}
		}

	}

	return mappings
}
