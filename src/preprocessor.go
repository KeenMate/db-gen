package dbGen

import (
	"github.com/stoewer/go-strcase"
	"log"
	"sort"
)

// Transforms data returned by database to structures that are used in generator

func Preprocess(routines []DbRoutine, config *Config) ([]Function, error) {
	// In future this should be more modular

	// Map routines
	functions, err := mapFunctions(routines, config)

	if err != nil {
		log.Println("Error while mapping functions")
		return nil, err
	}

	return functions, nil

}

func mapFunctions(routines []DbRoutine, config *Config) ([]Function, error) {
	mappedFunctions := make([]Function, len(routines))

	for i, routine := range routines {
		parameters := getParameters(routine.InParameters)
		functionName := getFunctionName(routine.RoutineName)
		dbFullFunctionName := routine.RoutineSchema + "." + routine.RoutineName
		modelName := getModelName(routine.RoutineName)
		processorName := getProcessorName(routine.RoutineName)
		returnProperties := getParameters(routine.OutParameters)

		function := &Function{
			FunctionName:       functionName,
			DbFullFunctionName: dbFullFunctionName,
			ModelName:          modelName,
			Parameters:         parameters,
			ReturnProperties:   returnProperties,
			ProcessorName:      processorName,
			HasReturn:          len(returnProperties) > 0,
			IsProcedure:        routine.FuncType == "procedure",
		}

		mappedFunctions[i] = *function
	}

	return mappedFunctions, nil
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
	case "boolean":
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
		return "byte[]", "GetValue"
	default:
		return "string", "GetString"
	}
}
