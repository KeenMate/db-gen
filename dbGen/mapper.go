package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	"github.com/stoewer/go-strcase"
	"slices"
	"sort"
)

type mapping struct {
	mappedFunction string
	mappedType     string
}

type effectiveParamMapping struct {
	name        string
	typeMapping mapping
	isNullable  bool
}

// TODO make configurable
const hiddenSchema = "public"

var structuredTypes = []string{"record", "USER-DEFINED"}
var voidTypes = []string{"void"}

var emptyMapping = RoutineMapping{
	Generate:            true,
	MappedName:          "",
	DontSelectValue:     false,
	SelectOnlySpecified: false,
	Model:               make(map[string]ColumnMapping),
	Parameters:          make(map[string]ParamMapping),
}

func mapFunctions(routines *[]DbRoutine, globalTypeMappings *map[string]mapping, config *Config) ([]Routine, error) {
	mappedFunctions := make([]Routine, len(*routines))
	schemaConfig := getSchemaConfigMap(config)

	for i, routine := range *routines {
		common.LogDebug("Mapping %s", routine.RoutineName)
		routineMapping := getRoutineMapping(routine, schemaConfig)

		returnProperties, err := mapModel(routine, globalTypeMappings, &routineMapping, config)
		if err != nil {
			return nil, fmt.Errorf("processing function %s: %s", routine.RoutineName, err)
		}

		parameters, err := mapParameters(routine.InParameters, globalTypeMappings, &routineMapping, config)
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

func mapModel(routine DbRoutine, globalTypeMappings *map[string]mapping, routineMapping *RoutineMapping, config *Config) ([]Property, error) {

	returnParameters := make([]Property, 0)

	//procedures in pg don't have return type
	if routine.FuncType == Procedure || slices.Contains(voidTypes, routine.DataType) {
		return returnParameters, nil
	}

	columns := routine.OutParameters

	// If value is simple data type
	if !slices.Contains(structuredTypes, routine.DataType) {

		// if function has return type, it means it return just one value
		columns = []DbParameter{{
			OrdinalPosition: 0,
			Name:            routine.RoutineName,
			Mode:            OutMode,
			UDTName:         routine.DataType,
			IsNullable:      false,
		}}

	}

	properties := make([]Property, 0)

	if columns == nil || len(columns) == 0 {
		return properties, nil
	}

	// Make sure attributes are in right order
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].OrdinalPosition < columns[j].OrdinalPosition
	})

	// position is relative to ordinal position of first column
	positionOffset := columns[0].OrdinalPosition

	for _, column := range columns {
		shouldSelect, columnMapping, err := getColumnMapping(column, routineMapping, globalTypeMappings, config)

		if err != nil {
			return nil, fmt.Errorf("getting effective mapping of %s: %s", column.Name, err)
		}

		if !shouldSelect {
			common.LogDebug("skipping selection of %s", column)
			continue
		}

		property := Property{
			DbColumnName:   column.Name,
			DbColumnType:   column.UDTName,
			PropertyName:   columnMapping.name,
			PropertyType:   columnMapping.typeMapping.mappedType,
			Position:       column.OrdinalPosition - positionOffset,
			MapperFunction: columnMapping.typeMapping.mappedFunction,
			Nullable:       columnMapping.isNullable,
		}

		properties = append(properties, property)
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
		effectiveMapping, err := getParamMapping(parameter, routineMapping, typeMappings, config)
		if err != nil {
			return nil, fmt.Errorf("processing parameter %s: %s", parameter.Name, err)
		}

		property := &Property{
			DbColumnName:   parameter.Name,
			DbColumnType:   parameter.UDTName,
			PropertyName:   effectiveMapping.name,
			PropertyType:   effectiveMapping.typeMapping.mappedType,
			Position:       parameter.OrdinalPosition - positionOffset,
			MapperFunction: "",
			Nullable:       effectiveMapping.isNullable,
		}

		properties[i] = *property
	}

	return properties, nil
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

func getColumnMapping(param DbParameter, routineMapping *RoutineMapping, globalMappings *map[string]mapping, config *Config) (bool, *effectiveParamMapping, error) {
	if routineMapping.DontSelectValue {
		return false, nil, nil
	}

	name := strcase.UpperCamelCase(param.Name)
	isNullable := param.IsNullable
	var typeMapping *mapping = nil
	var err error = nil

	explicitMapping, hasExplicitParamMapping := routineMapping.Model[param.Name]
	if !hasExplicitParamMapping && routineMapping.SelectOnlySpecified {
		return false, nil, nil
	}

	if hasExplicitParamMapping {
		if !explicitMapping.SelectColumn {
			return false, nil, nil
		}

		if explicitMapping.MappedName != "" {
			name = explicitMapping.MappedName
		}

		if explicitMapping.IsNullable.Valid {
			isNullable = explicitMapping.IsNullable.Bool
		}

		if explicitMapping.MappedType != "" {
			typeMapping, err = handleTypeMappingOverride(explicitMapping.MappedType, "", config)
			if err != nil {
				return false, nil, err
			}

		}
	}

	if typeMapping == nil {
		typeMapping, err = getTypeMapping(param.UDTName, globalMappings)
		if err != nil {
			return false, nil, err
		}
	}

	return true, &effectiveParamMapping{
		name:        name,
		typeMapping: *typeMapping,
		isNullable:  isNullable,
	}, nil

}

func getParamMapping(param DbParameter, routineMapping *RoutineMapping, globalMappings *map[string]mapping, config *Config) (*effectiveParamMapping, error) {
	name := param.Name
	isNullable := param.IsNullable
	var typeMapping *mapping = nil
	var err error = nil

	explicitMapping, hasExplicitParamMapping := routineMapping.Parameters[param.Name]
	if hasExplicitParamMapping {
		if explicitMapping.MappedName != "" {
			name = explicitMapping.MappedName
		}

		if explicitMapping.IsNullable.Valid {
			isNullable = explicitMapping.IsNullable.Bool
		}

		if explicitMapping.MappedType != "" {
			typeMapping, err = handleTypeMappingOverride(explicitMapping.MappedType, "", config)
			if err != nil {
				return nil, err
			}

		}
	}

	if typeMapping == nil {
		typeMapping, err = getTypeMapping(param.UDTName, globalMappings)
		if err != nil {
			return nil, err
		}
	}

	return &effectiveParamMapping{
		name:        name,
		typeMapping: *typeMapping,
		isNullable:  isNullable,
	}, nil

}

func getFunctionName(dbFunctionName string, schema string, mappedName string) string {
	if mappedName != "" {
		return mappedName
	}

	// If you want to use different case, use template function in templates

	schemaPrefix := ""
	// don't add public_ to function names
	if schema != hiddenSchema {
		schemaPrefix = strcase.UpperCamelCase(schema)
	}
	return schemaPrefix + strcase.UpperCamelCase(dbFunctionName)
}

func getModelName(functionName string) string {
	return strcase.UpperCamelCase(functionName) + "Model"
}
func getProcessorName(functionName string) string {
	return strcase.UpperCamelCase(functionName) + "Processor"
}

func getTypeMapping(dbDataType string, mappings *map[string]mapping) (*mapping, error) {
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

func handleTypeMappingOverride(typeOverride string, mappingFunctionOverride string, config *Config) (*mapping, error) {
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
