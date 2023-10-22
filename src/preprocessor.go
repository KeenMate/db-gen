package dbGen

import (
	"fmt"
	"github.com/stoewer/go-strcase"
	"log"
	"slices"
	"sort"
)

// Transforms data returned by database to structures that are used in generator

func Preprocess(routines []DbRoutine, config *Config) ([]Routine, error) {
	// In future this should be more modular

	filteredFunctions := filterFunctions(&routines, config)

	// don't need to compute for every property
	typeMappings := getTypeMappings(config)
	VerboseLog("Got %d type mappigns", len(typeMappings))

	// Map routines
	functions, err := mapFunctions(&filteredFunctions, &typeMappings, config)

	if err != nil {
		log.Println("Error while mapping functions")
		return nil, err
	}

	return functions, nil

}

func filterFunctions(functions *[]DbRoutine, config *Config) []DbRoutine {
	schemaMap := getSchemaConfigMap(config)
	VerboseLog("Got %d schema configs  ", len(schemaMap))
	filteredFunctions := make([]DbRoutine, 0)

	for _, function := range *functions {
		schemaConfig, exists := schemaMap[function.RoutineSchema]

		// if config for given schema doest exits, don't generate for any function in given scheme
		if !exists {
			VerboseLog("No schema config for '%s'", function.RoutineSchema)
			continue
		}

		if schemaConfig.AllFunctions || slices.Contains(schemaConfig.Functions, function.RoutineName) {
			// Case sensitive
			if slices.Contains(schemaConfig.IgnoredFunctions, function.RoutineName) {
				VerboseLog("Routine '%s.%s' in ignored functions", function.RoutineSchema, function.RoutineName)
				continue
			}

			filteredFunctions = append(filteredFunctions, function)
		} else {
			VerboseLog("Routine '%s.%s' not generated because all function is false or isnt included in functions",
				function.RoutineSchema,
				function.RoutineName)
		}

	}

	return filteredFunctions
}

func getSchemaConfigMap(config *Config) map[string]SchemaConfig {
	schemaMap := make(map[string]SchemaConfig)

	for _, schemaConfig := range config.Generate {
		schemaMap[schemaConfig.Schema] = schemaConfig
	}

	return schemaMap
}

func mapFunctions(routines *[]DbRoutine, typeMappings *map[string]mapping, config *Config) ([]Routine, error) {
	mappedFunctions := make([]Routine, len(*routines))

	for i, routine := range *routines {
		VerboseLog("Mapping %s", routine.RoutineName)

		returnProperties, err := getReturnProperties(routine, typeMappings)
		if err != nil {
			return nil, fmt.Errorf("mapping function %s: %s", routine.RoutineName, err)
		}

		parameters, err := getParameters(routine.InParameters, typeMappings)
		if err != nil {
			return nil, fmt.Errorf("mapping function %s: %s", routine.RoutineName, err)
		}

		functionName := getFunctionName(routine.RoutineName)
		dbFullFunctionName := routine.RoutineSchema + "." + routine.RoutineName
		modelName := getModelName(routine.RoutineName)
		processorName := getProcessorName(routine.RoutineName)

		function := &Routine{
			FunctionName:       functionName,
			DbFullFunctionName: dbFullFunctionName,
			ModelName:          modelName,
			Parameters:         parameters,
			ReturnProperties:   returnProperties,
			ProcessorName:      processorName,
			HasReturn:          len(returnProperties) > 0,
			IsProcedure:        routine.FuncType == Procedure,
			Schema:             routine.RoutineSchema,
			DbFunctionName:     routine.RoutineName,
		}

		mappedFunctions[i] = *function
	}

	return mappedFunctions, nil
}

func getReturnProperties(routine DbRoutine, typeMappings *map[string]mapping) ([]Property, error) {

	returnParameters := make([]Property, 0)
	structuredTypes := []string{"record", "USER-DEFINED"}
	voidTypes := []string{"void"}

	//procedures in pg don't have return type
	if routine.FuncType == Procedure || slices.Contains(voidTypes, routine.DataType) {
		return returnParameters, nil
	}

	outParameters := routine.OutParameters

	// If value is simple data type
	if !slices.Contains(structuredTypes, routine.DataType) {

		// TODO investigate this more
		// if function has return type, it means it return just one value
		outParameters = []DbParameter{{
			OrdinalPosition: 0,
			Name:            routine.RoutineName,
			Mode:            OutMode,
			UDTName:         routine.DataType,
			IsNullable:      false,
		}}

	}

	return getParameters(outParameters, typeMappings)
}

func getParameters(attributes []DbParameter, typeMappings *map[string]mapping) ([]Property, error) {

	properties := make([]Property, len(attributes))

	if attributes == nil || len(attributes) == 0 {
		return properties, nil
	}

	// Make sure attributes are in right order
	sort.Slice(attributes, func(i, j int) bool {
		return attributes[i].OrdinalPosition < attributes[j].OrdinalPosition
	})

	// First possition should be 0
	positionOffset := attributes[0].OrdinalPosition
	//VerboseLog("Possition offset is %d", positionOffset)

	for i, attribute := range attributes {
		propertyName := getPropertyName(attribute.Name)
		typeMapping, err := getMapping(typeMappings, attribute.UDTName)
		if err != nil {
			return nil, fmt.Errorf("mapping parameter %s: %s", attribute.Name, err)
		}

		property := &Property{
			DbColumnName:   attribute.Name,
			DbColumnType:   attribute.UDTName,
			PropertyName:   propertyName,
			PropertyType:   typeMapping.mappingType,
			Position:       attribute.OrdinalPosition - positionOffset,
			MapperFunction: typeMapping.mappingFunction,
		}

		properties[i] = *property
	}

	return properties, nil
}

// If you want to use different case, use template function in templates
func getFunctionName(dbFunctionName string) string {
	return strcase.UpperCamelCase(dbFunctionName)
}

func getPropertyName(dbColumnName string) string {
	return strcase.UpperCamelCase(dbColumnName)
}
func getModelName(dbColumnName string) string {
	return strcase.UpperCamelCase(dbColumnName) + "Model"
}
func getProcessorName(dbColumnName string) string {
	return strcase.UpperCamelCase(dbColumnName) + "Processor"
}

type mapping struct {
	mappingFunction string
	mappingType     string
}

func getTypeMappings(config *Config) map[string]mapping {
	mappings := make(map[string]mapping)

	// If there are multiple mappings to one database type, last one will be used

	for _, val := range config.Mappings {
		for _, databaseType := range val.DatabaseTypes {
			mappings[databaseType] = mapping{
				mappingFunction: val.MappingFunction,
				mappingType:     val.MappedType,
			}
		}

	}

	return mappings
}

func getMapping(mappings *map[string]mapping, dbDataType string) (*mapping, error) {
	val, isFound := (*mappings)[dbDataType]

	if !isFound {
		return nil, fmt.Errorf("mapping for dbType '%s' not found", dbDataType)
	}

	return &val, nil
}
