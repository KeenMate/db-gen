package dbGen

import (
	"fmt"
	"github.com/guregu/null/v5"
	"github.com/keenmate/db-gen/common"
	"github.com/stoewer/go-strcase"
	"slices"
	"sort"
)

// TODO make configurable
const hiddenSchema = "public"

func mapFunctions(routines *[]DbRoutine, typeMappings *map[string]mapping, config *Config) ([]Routine, error) {
	mappedFunctions := make([]Routine, len(*routines))
	schemaConfig := getSchemaConfigMap(config)

	for i, routine := range *routines {
		common.LogDebug("Mapping %s", routine.RoutineName)
		routineMapping := getRoutineMapping(routine, schemaConfig)

		returnProperties, err := mapReturnColumns(routine, typeMappings, &routineMapping, config)
		if err != nil {
			return nil, fmt.Errorf("processing function %s: %s", routine.RoutineName, err)
		}

		parameters, err := mapParameters(routine.InParameters, typeMappings, &routineMapping, config)
		if err != nil {
			return nil, fmt.Errorf("processing function %s: %s", routine.RoutineName, err)
		}

		functionName := getFunctionName(routine.RoutineName, routine.RoutineSchema, routineMapping.MappedName)

		modelName := getModelName(functionName)
		processorName := getProcessorName(functionName)

		function := &Routine{
			FunctionName:       functionName,
			DbFullFunctionName: routine.RoutineSchema + "." + routine.RoutineName,
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

func mapReturnColumns(routine DbRoutine, typeMappings *map[string]mapping, routineMapping *RoutineMapping, config *Config) ([]Property, error) {

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

	properties := make([]Property, 0)

	if outParameters == nil || len(outParameters) == 0 {
		return properties, nil
	}

	// Make sure attributes are in right order
	sort.Slice(outParameters, func(i, j int) bool {
		return outParameters[i].OrdinalPosition < outParameters[j].OrdinalPosition
	})

	// First possition should be 0
	positionOffset := outParameters[0].OrdinalPosition
	//common.LogDebug("Possition offset is %d", positionOffset)

	for _, column := range outParameters {
		shouldSelect, columnMapping := getColumnMapping(column.Name, routineMapping)

		if !shouldSelect {
			common.LogDebug("skipping selection of %s", column)
			continue
		}

		propertyName := getPropertyName(column.Name, columnMapping.MappedName)

		typeMapping, err := getTypeMapping(column.UDTName, columnMapping.MappedType, columnMapping.MappingFunction, typeMappings, config)
		if err != nil {
			return nil, fmt.Errorf("processing parameter %s: %s", column.Name, err)
		}

		property := &Property{
			DbColumnName:   column.Name,
			DbColumnType:   column.UDTName,
			PropertyName:   propertyName,
			PropertyType:   typeMapping.mappedType,
			Position:       column.OrdinalPosition - positionOffset,
			MapperFunction: typeMapping.mappedFunction,
			Nullable:       getIsNullable(columnMapping.IsNullable, column.IsNullable),
		}

		properties = append(properties, *property)
	}

	return properties, nil
}

func mapParameters(attributes []DbParameter, typeMappings *map[string]mapping, routineMapping *RoutineMapping, config *Config) ([]Property, error) {

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
	//common.LogDebug("Possition offset is %d", positionOffset)

	for i, parameter := range attributes {
		paramMapping := getParamMapping(parameter, routineMapping)
		common.LogDebug("parameter %s custom mapping: %+v", parameter.Name, paramMapping)
		propertyName := getPropertyName(parameter.Name, paramMapping.MappedName)

		typeMapping, err := getTypeMapping(parameter.UDTName, paramMapping.MappedType, "", typeMappings, config)
		if err != nil {
			return nil, fmt.Errorf("processing parameter %s: %s", parameter.Name, err)
		}

		property := &Property{
			DbColumnName:   parameter.Name,
			DbColumnType:   parameter.UDTName,
			PropertyName:   propertyName,
			PropertyType:   typeMapping.mappedType,
			Position:       parameter.OrdinalPosition - positionOffset,
			MapperFunction: "",
			Nullable:       getIsNullable(paramMapping.IsNullable, parameter.IsNullable),
		}

		properties[i] = *property
	}

	return properties, nil
}

// If you want to use different case, use template function in templates
func getFunctionName(dbFunctionName string, schema string, mappedName string) string {
	if mappedName != "" {
		return mappedName
	}

	schemaPrefix := ""
	// don't add public_ to function names
	if schema != hiddenSchema {
		schemaPrefix = strcase.UpperCamelCase(schema)
	}
	return schemaPrefix + strcase.UpperCamelCase(dbFunctionName)
}

func getPropertyName(dbColumnName string, mappedName string) string {
	if mappedName != "" {
		return mappedName
	}

	return strcase.UpperCamelCase(dbColumnName)
}
func getModelName(functionName string) string {
	return strcase.UpperCamelCase(functionName) + "Model"
}
func getProcessorName(functionName string) string {
	return strcase.UpperCamelCase(functionName) + "Processor"
}

type mapping struct {
	mappedFunction string
	mappedType     string
}

func getTypeMapping(dbDataType string, typeOverride string, mappingFunctionOverride string, mappings *map[string]mapping, config *Config) (*mapping, error) {
	// mapped type override
	if typeOverride != "" {
		return handleMappingOverride(typeOverride, mappingFunctionOverride, config)
	}

	val, isFound := (*mappings)[dbDataType]

	if !isFound {
		fallbackVal, fallbackExists := (*mappings)["*"]

		if !fallbackExists {
			return nil, fmt.Errorf("processing for dbType '%s' not found and fallback processing * is not set ", dbDataType)

		}

		common.LogDebug("Using fallback value %+v for type %s", fallbackVal, dbDataType)

		return &fallbackVal, nil
	}

	return &val, nil
}

func handleMappingOverride(typeOverride string, mappingFunctionOverride string, config *Config) (*mapping, error) {
	if mappingFunctionOverride != "" {
		return &mapping{
			mappedFunction: mappingFunctionOverride,
			mappedType:     typeOverride,
		}, nil
	}

	// get mapping function
	for _, typeMapping := range config.Mappings {
		if typeMapping.MappedType == typeOverride {
			return &mapping{
				mappedFunction: typeMapping.MappingFunction,
				mappedType:     typeOverride,
			}, nil
		}
	}

	// no mapping function is set and no mapping exist for type given
	return nil, fmt.Errorf("mapped type overriden to %s, but no mapping functions specified and mapping function for override type doenst exist in mappings", typeOverride)
}

var emptyMapping = RoutineMapping{
	Generate:            true,
	MappedName:          "",
	DontSelectValue:     false,
	SelectOnlySpecified: false,
	Model:               make(map[string]ColumnMapping),
	Parameters:          make(map[string]ParamMapping),
}

func getRoutineMapping(routine DbRoutine, schemaConfigs map[string]SchemaConfig) RoutineMapping {
	schemaConfig, ok := schemaConfigs[routine.RoutineSchema]
	if !ok {
		// this should never happen
		panic("trying ty get function mapping for function in schema that is not defined. This should never happen, because function should have been fitered out")
	}
	routineMapping, found := schemaConfig.Functions[routine.RoutineNameWithParams]
	if found {
		return routineMapping
	}

	routineMapping, found = schemaConfig.Functions[routine.RoutineName]
	if found {
		return routineMapping
	}

	return emptyMapping
}

var emptyColumnMapping = ColumnMapping{
	SelectColumn:    true,
	MappedName:      "",
	MappedType:      "",
	MappingFunction: "",
	IsNullable:      null.NewBool(false, false),
}

func getColumnMapping(columnName string, routineMapping *RoutineMapping) (bool, ColumnMapping) {
	if routineMapping.DontSelectValue {
		return false, emptyColumnMapping
	}

	columnMapping, hasExplicitColumnMapping := routineMapping.Model[columnName]
	if hasExplicitColumnMapping {
		common.LogDebug("explicit mapping on column %s", columnName)
		return columnMapping.SelectColumn, columnMapping
	}

	return !routineMapping.SelectOnlySpecified, emptyColumnMapping

}

var emptyParamMapping = ParamMapping{
	MappedName: "",
	MappedType: "",
	IsNullable: null.NewBool(false, false),
}

func getParamMapping(param DbParameter, routineMapping *RoutineMapping) ParamMapping {

	columnMapping, hasExplicitParamMapping := routineMapping.Parameters[param.Name]
	if hasExplicitParamMapping {
		return columnMapping
	}

	return emptyParamMapping

}

func getIsNullable(typeMappingOverride null.Bool, implicitValue bool) bool {
	if typeMappingOverride.Valid {
		common.LogDebug("using explicit value of nullable %t", typeMappingOverride.Bool)
		return typeMappingOverride.Bool
	}

	return implicitValue
}
