package dbGen

import (
	"fmt"
	"github.com/stoewer/go-strcase"
	"log"
	"slices"
	"sort"
)

// Transforms data returned by database to structures that are used in generator

func Preprocess(routines []DbRoutine, config *Config) ([]Function, error) {
	// In future this should be more modular

	// Map routines
	functions, err := mapFunctions(&routines, config)

	if err != nil {
		log.Println("Error while mapping functions")
		return nil, err
	}

	functions = filterFunctions(&functions, config)

	return functions, nil

}

func filterFunctions(functions *[]Function, config *Config) []Function {
	schemaMap := getSchemaConfigMap(config)

	filteredFunctions := make([]Function, 0)

	for _, function := range *functions {
		schemaConfig, exists := schemaMap[function.Schema]

		// if config for given schema doest exits, dont generate for any function in given scheme
		if !exists {
			continue
		}

		if schemaConfig.AllFunctions || slices.Contains(schemaConfig.Functions, function.FunctionName) {
			// Case sensitive
			if slices.Contains(schemaConfig.IgnoredFunctions, function.FunctionName) {
				VerboseLog(fmt.Sprintf("Ignoring funtion '%s' in scheme '%s'", function.FunctionName, function.Schema))
				continue
			}

			filteredFunctions = append(filteredFunctions, function)

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

func mapFunctions(routines *[]DbRoutine, config *Config) ([]Function, error) {
	mappedFunctions := make([]Function, len(*routines))

	for i, routine := range *routines {
		parameters := getParameters(routine.InParameters)
		functionName := getFunctionName(routine.RoutineName)
		dbFullFunctionName := routine.RoutineSchema + "." + routine.RoutineName
		modelName := getModelName(routine.RoutineName)
		processorName := getProcessorName(routine.RoutineName)
		returnProperties := getReturnProperties(routine)

		function := &Function{
			FunctionName:       functionName,
			DbFullFunctionName: dbFullFunctionName,
			ModelName:          modelName,
			Parameters:         parameters,
			ReturnProperties:   returnProperties,
			ProcessorName:      processorName,
			HasReturn:          len(returnProperties) > 0,
			IsProcedure:        routine.FuncType == "procedure",
			Schema:             routine.RoutineSchema,
		}

		mappedFunctions[i] = *function
	}

	return mappedFunctions, nil
}

func getReturnProperties(routine DbRoutine) []Property {
	returnParameters := make([]Property, 0)
	structuredTypes := []string{"record", "USER-DEFINED"}
	voidTypes := []string{"void"}

	if slices.Contains(voidTypes, routine.DataType) {
		return returnParameters
	}

	// If value is simple data type
	if !slices.Contains(structuredTypes, routine.DataType) {

		propertyName := getPropertyName(routine.RoutineName)
		propertyType, propertyMapper := getCsharpType(routine.DataType)

		return append(returnParameters, Property{
			DbColumnName:   routine.RoutineName,
			DbColumnType:   routine.DataType,
			PropertyName:   propertyName,
			PropertyType:   propertyType,
			Position:       0,
			MapperFunction: propertyMapper,
		})
	}

	return getParameters(routine.OutParameters)
}

func getParameters(attributes []DbParameter) []Property {
	properties := make([]Property, len(attributes))

	// Make sure attributes are in right order
	sort.Slice(attributes, func(i, j int) bool {
		return attributes[i].OrdinalPosition < attributes[j].OrdinalPosition
	})

	for i, attribute := range attributes {
		propertyName := getPropertyName(attribute.Name)
		propertyType, propertyMapper := getCsharpType(attribute.UDTName)

		property := &Property{
			DbColumnName:   attribute.Name,
			DbColumnType:   attribute.UDTName,
			PropertyName:   propertyName,
			PropertyType:   propertyType,
			Position:       attribute.OrdinalPosition - 1,
			MapperFunction: propertyMapper,
		}

		properties[i] = *property
	}

	return properties
}

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

func getCsharpType(databaseType string) (string, string) {
	// TODO this is hardcodded for now
	// But in future it should be loaded from file
	switch databaseType {
	case "boolean", "bool":
		return "bool", "GetBoolean"
	case "smallint", "int2":
		return "short", "GetInt16"
	case "integer", "int4":
		return "int", "GetInt32"
	case "bigint", "int8":
		return "long", "GetInt64"
	case "real":
		return "float", "GetFloat"
	case "double precision":
		return "double", "GetDouble"
	case "numeric", "money":
		return "decimal", "GetDecimal"
	case "text", "character varying", "character", "citext", "json", "jsonb", "xml":
		return "string", "GetString"
	case "uuid":
		return "Guid", "GetGuid"
	case "bytea":
		return "byte[]", "GetByteArray"
	case "timestamptz", "date", "timestamp":
		return "DateTime", "GetDateTime"
	case "interval":
		return "TimeSpan", "GetTimeSpan"
	default:
		return "string", "GetString"
	}
}
