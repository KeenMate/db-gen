package dbGen

import (
	"fmt"
	common2 "github.com/keenmate/db-gen/private/helpers"
	"slices"
	"sort"
	"strings"
)

type mapping struct {
	mappedFunction        string
	mappedType            string
	nullableReturnType    string
	nullableParameterType string
	optionalParameterType string
}

type effectiveParamMapping struct {
	name          string
	typeMapping   mapping
	isNullable    bool
	isOptional    bool
	securityLevel string
}

// Parameter logging-sensitivity levels. See docs/configuration.md.
const (
	SecurityLevelNone   = "none"   // safe to log as plain text
	SecurityLevelSecure = "secure" // plain text or masked depending on a runtime insecure-logging flag
	SecurityLevelStrict = "strict" // always masked
	SecurityLevelOmit   = "omit"   // never logged
)

// DefaultSecurityLevel is the fallback level when nothing else matches a parameter.
const DefaultSecurityLevel = SecurityLevelSecure

var ValidSecurityLevels = []string{SecurityLevelNone, SecurityLevelSecure, SecurityLevelStrict, SecurityLevelOmit}

// TODO make configurable
const hiddenSchema = "public"

const fallbackMappingKey = "*"

// this data types represent structured data
var structuredTypes = []string{"record", "USER-DEFINED"}

// data types that represents no return types
var voidTypes = []string{"void"}

var emptyMapping = RoutineMapping{
	Generate:            true,
	MappedName:          "",
	DontRetrieveValues:  false,
	SelectOnlySpecified: false,
	Model:               make(map[string]ColumnMapping),
	Parameters:          make(map[string]ParamMapping),
}

func processContextParameters(parameters []Property, config *Config) (contextParams []Property, regularParams []Property, allParams []Property) {
	contextParams = make([]Property, 0)
	regularParams = make([]Property, 0)
	allParams = make([]Property, len(parameters))

	// Build a quick lookup map for context parameter mappings
	contextParamMap := make(map[string]string)
	for _, mapping := range config.ContextParameterMappings {
		for _, paramName := range mapping.ParameterNames {
			contextParamMap[strings.ToLower(paramName)] = mapping.ContextPath
		}
	}

	// Process parameters in original order, preserving order in allParams
	for i, param := range parameters {
		paramNameLower := strings.ToLower(param.PropertyName)
		if contextPath, isContext := contextParamMap[paramNameLower]; isContext {
			// Mark as context parameter
			param.IsContextParameter = true
			param.ContextPath = contextPath
			contextParams = append(contextParams, param)
		} else {
			regularParams = append(regularParams, param)
		}
		allParams[i] = param
	}

	return contextParams, regularParams, allParams
}

func mapRoutines(routines *[]DbRoutine, globalTypeMappings *map[string]mapping, config *Config) ([]Routine, error) {
	mappedFunctions := make([]Routine, len(*routines))
	schemaConfig := getSchemaConfigMap(config)

	for i, routine := range *routines {
		common2.LogDebug("Mapping %s", routine.RoutineName)
		routineMapping := getRoutineMapping(routine, schemaConfig)

		modelProperties, err := mapModel(routine, globalTypeMappings, &routineMapping, config)
		if err != nil {
			return nil, fmt.Errorf("processing function %s: %s", routine.RoutineName, err)
		}

		parameters, err := mapParameters(routine.InParameters, globalTypeMappings, &routineMapping, config)
		if err != nil {
			return nil, fmt.Errorf("processing function %s: %s", routine.RoutineName, err)
		}

		// Process context parameter mappings
		contextParams, regularParams, allParameters := processContextParameters(parameters, config)
		usesUserContext := len(contextParams) > 0 && config.UseUserContext

		// Process validation rules
		processParameterValidations(allParameters, routine, config)

		// default case for names is UpperCamelcase
		functionName := getFunctionName(routine.RoutineName, routine.RoutineSchema, routineMapping.MappedName, routine.OverloadSuffix)
		modelName := getModelName(functionName)
		processorName := getProcessorName(functionName)

		mappedRoutine := Routine{
			FunctionName:       functionName,
			DbFullFunctionName: routine.RoutineSchema + "." + routine.RoutineName,
			ModelName:          modelName,
			Parameters:         allParameters,
			ReturnProperties:   modelProperties,
			ProcessorName:      processorName,
			HasReturn:          len(modelProperties) > 0,
			IsProcedure:        routine.FuncType == Procedure,
			Schema:             routine.RoutineSchema,
			DbFunctionName:     routine.RoutineName,
			UsesUserContext:    usesUserContext,
			ContextParameters:  contextParams,
			RegularParameters:  regularParams,
		}

		mappedFunctions[i] = mappedRoutine
	}

	return mappedFunctions, nil
}

func mapModel(routine DbRoutine, globalTypeMappings *map[string]mapping, routineMapping *RoutineMapping, config *Config) ([]Property, error) {

	modelProperties := make([]Property, 0)

	//procedures in pg don't have return type
	if routine.FuncType == Procedure || slices.Contains(voidTypes, routine.DataType) {
		return modelProperties, nil
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
			common2.LogDebug("skipping selection of %s", column)
			continue
		}

		// Determine the actual PropertyType for return values based on nullable status
		propertyType := columnMapping.typeMapping.mappedType
		if columnMapping.isNullable && columnMapping.typeMapping.nullableReturnType != "" {
			propertyType = columnMapping.typeMapping.nullableReturnType
		}

		property := Property{
			DbColumnName:       column.Name,
			DbColumnType:       column.UDTName,
			PropertyName:       columnMapping.name,
			PropertyType:       propertyType,
			BaseType:           columnMapping.typeMapping.mappedType,
			NullableReturnType: columnMapping.typeMapping.nullableReturnType,
			NullableParamType:  columnMapping.typeMapping.nullableParameterType,
			OptionalParamType:  columnMapping.typeMapping.optionalParameterType,
			Position:           column.OrdinalPosition - positionOffset,
			MapperFunction:     columnMapping.typeMapping.mappedFunction,
			Nullable:           columnMapping.isNullable,
			Optional:           columnMapping.isOptional,
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
	//helpers.LogDebug("Possition offset is %d", positionOffset)

	for i, parameter := range attributes {
		effectiveMapping, err := getParamMapping(parameter, routineMapping, typeMappings, config)
		if err != nil {
			return nil, fmt.Errorf("processing parameter %s: %s", parameter.Name, err)
		}

		// Determine the actual PropertyType based on nullable/optional status
		propertyType := effectiveMapping.typeMapping.mappedType
		if effectiveMapping.isOptional && effectiveMapping.typeMapping.optionalParameterType != "" {
			propertyType = effectiveMapping.typeMapping.optionalParameterType
		} else if effectiveMapping.isNullable && effectiveMapping.typeMapping.nullableParameterType != "" {
			propertyType = effectiveMapping.typeMapping.nullableParameterType
		}

		property := &Property{
			DbColumnName:       parameter.Name,
			DbColumnType:       parameter.UDTName,
			PropertyName:       effectiveMapping.name,
			PropertyType:       propertyType,
			BaseType:           effectiveMapping.typeMapping.mappedType,
			NullableReturnType: effectiveMapping.typeMapping.nullableReturnType,
			NullableParamType:  effectiveMapping.typeMapping.nullableParameterType,
			OptionalParamType:  effectiveMapping.typeMapping.optionalParameterType,
			Position:           parameter.OrdinalPosition - positionOffset,
			MapperFunction:     "",
			Nullable:           effectiveMapping.isNullable,
			Optional:           effectiveMapping.isOptional,
			SecurityLevel:      effectiveMapping.securityLevel,
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
	if routineMapping.DontRetrieveValues {
		return false, nil, nil
	}

	name := common2.ToPascalCase(param.Name)
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
			typeMapping, err = handleTypeMappingOverride(explicitMapping.MappedType, explicitMapping.MappingFunction, config)
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
		// no column has to be selected => column is always optional
		isOptional: true,
	}, nil

}

func getParamMapping(param DbParameter, routineMapping *RoutineMapping, globalMappings *map[string]mapping, config *Config) (*effectiveParamMapping, error) {
	name := param.Name
	isNullable := param.IsNullable
	isOptional := param.IsOptional
	var typeMapping *mapping = nil
	var err error = nil

	explicitMapping, hasExplicitParamMapping := routineMapping.Parameters[param.Name]
	securityLevel := resolveSecurityLevel(param.Name, explicitMapping, hasExplicitParamMapping, config)
	if hasExplicitParamMapping {
		if explicitMapping.MappedName != "" {
			name = explicitMapping.MappedName
		}

		if explicitMapping.IsNullable.Valid {
			isNullable = explicitMapping.IsNullable.Bool
		}

		if explicitMapping.IsOptional.Valid {
			isOptional = explicitMapping.IsOptional.Bool
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
		name:          name,
		typeMapping:   *typeMapping,
		isNullable:    isNullable,
		isOptional:    isOptional,
		securityLevel: securityLevel,
	}, nil

}

// resolveSecurityLevel determines a parameter's logging-sensitivity level.
// Precedence: per-function override > global by-name mapping > DefaultParameterSecurityLevel.
// All configured values are already normalized to lowercase in normalizeAndValidateSecurityLevels.
func resolveSecurityLevel(dbParamName string, explicitMapping ParamMapping, hasExplicitMapping bool, config *Config) string {
	if hasExplicitMapping && explicitMapping.SecurityLevel != "" {
		return explicitMapping.SecurityLevel
	}

	nameLower := strings.ToLower(dbParamName)
	for _, mapping := range config.ParameterSecurityMappings {
		for _, paramName := range mapping.ParameterNames {
			if strings.ToLower(paramName) == nameLower {
				return mapping.SecurityLevel
			}
		}
	}

	return config.DefaultParameterSecurityLevel
}

// normalizeAndValidateSecurityLevels lowercases every configured security level and
// rejects unknown values. Empty values are left as-is (they mean "not set" and fall
// through to the next precedence tier).
func normalizeAndValidateSecurityLevels(config *Config) error {
	validate := func(where, level string) (string, error) {
		if level == "" {
			return level, nil
		}
		normalized := strings.ToLower(level)
		if !common2.Contains(ValidSecurityLevels, normalized) {
			return "", fmt.Errorf("'%s' is not a valid security level for %s (valid: %s)",
				level, where, strings.Join(ValidSecurityLevels, ", "))
		}
		return normalized, nil
	}

	var err error
	if config.DefaultParameterSecurityLevel, err = validate("DefaultParameterSecurityLevel", config.DefaultParameterSecurityLevel); err != nil {
		return err
	}
	if config.DefaultParameterSecurityLevel == "" {
		config.DefaultParameterSecurityLevel = DefaultSecurityLevel
	}

	for i := range config.ParameterSecurityMappings {
		m := &config.ParameterSecurityMappings[i]
		if m.SecurityLevel, err = validate("ParameterSecurityMappings", m.SecurityLevel); err != nil {
			return err
		}
	}

	for si := range config.Generate {
		for fnName, routineMapping := range config.Generate[si].Functions {
			for pName, paramMapping := range routineMapping.Parameters {
				if paramMapping.SecurityLevel, err = validate(
					fmt.Sprintf("Generate[%s].Parameters[%s]", fnName, pName), paramMapping.SecurityLevel); err != nil {
					return err
				}
				routineMapping.Parameters[pName] = paramMapping
			}
		}
	}

	return nil
}

func getFunctionName(dbFunctionName string, schema string, mappedName string, overloadSuffix string) string {
	// An explicit MappedName wins outright — the caller has chosen the name and
	// is responsible for keeping overloads unique, so we don't append a suffix.
	if mappedName != "" {
		return mappedName
	}

	// If you want to use different case, use template function in templates

	schemaPrefix := ""
	// don't add public_ to function names
	if schema != hiddenSchema {
		schemaPrefix = common2.ToPascalCase(common2.NormalizeStr(schema))
	}
	// overloadSuffix is empty for non-overloaded routines; for overloads it is a
	// stable 1-based index that disambiguates the otherwise-identical names.
	return schemaPrefix + common2.ToPascalCase(dbFunctionName) + overloadSuffix
}

func getModelName(functionName string) string {
	return common2.ToPascalCase(functionName) + "Model"
}
func getProcessorName(functionName string) string {
	return common2.ToPascalCase(functionName) + "Processor"
}

// getTypeMapping if explicit mapping doesnt exist, try fallback
func getTypeMapping(dbDataType string, globalTypesMappings *map[string]mapping) (*mapping, error) {
	val, specificMappingExists := (*globalTypesMappings)[dbDataType]

	if !specificMappingExists {
		fallbackVal, fallbackExists := (*globalTypesMappings)[fallbackMappingKey]

		if !fallbackExists {
			return nil, fmt.Errorf("processing for dbType '%s' not found and fallback processing * is not set ", dbDataType)

		}

		common2.LogDebug("Using fallback value %+v for type %s", fallbackVal, dbDataType)

		return &fallbackVal, nil
	}

	return &val, nil
}

// handleTypeMappingOverride used when parsing model and parameters when mappedType is set
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
				mappedFunction:        typeMapping.MappingFunction,
				mappedType:            typeOverride,
				nullableReturnType:    typeMapping.NullableReturnType,
				nullableParameterType: typeMapping.NullableParameterType,
				optionalParameterType: typeMapping.OptionalParameterType,
			}, nil
		}
	}

	// no mapping function is set and no mapping exist for type given
	return nil, fmt.Errorf("mapped type overriden to %s, but no mapping functions specified and mapping function for override type doenst exist in mappings", typeOverride)
}

// processParameterValidations adds validation rules to parameters
func processParameterValidations(parameters []Property, routine DbRoutine, config *Config) {
	if len(config.Validation.ValidationRuleDefinitions) == 0 {
		return
	}

	// Build lookup maps
	ruleDefinitions := buildRuleDefinitionsMap(config.Validation.ValidationRuleDefinitions)
	paramValidationMap := buildParamValidationMap(config.Validation.ParameterValidationMappings)

	for i := range parameters {
		param := &parameters[i]
		paramNameLower := strings.ToLower(param.PropertyName)

		// Start with empty validation rules
		param.ValidationRules = make([]ValidationRule, 0)

		// 1. Check global parameter validation mappings
		if rules, found := paramValidationMap[paramNameLower]; found {
			param.ValidationRules = append(param.ValidationRules,
				resolveValidationRules(rules, ruleDefinitions)...)
		}

		// 2. Check function-specific validations (override/extend global)
		if funcValidations, found := config.Validation.FunctionSpecificValidations[routine.RoutineName]; found {
			if paramRules, found := funcValidations[param.DbColumnName]; found {
				param.ValidationRules = append(param.ValidationRules,
					resolveValidationRules(paramRules, ruleDefinitions)...)
			}
		}
	}
}

func buildRuleDefinitionsMap(definitions []ValidationRuleDefinition) map[string]*ValidationRuleDefinition {
	result := make(map[string]*ValidationRuleDefinition)
	for i := range definitions {
		def := &definitions[i]
		result[strings.ToLower(def.Name)] = def
	}
	return result
}

func buildParamValidationMap(mappings []ParameterValidationMapping) map[string][]interface{} {
	result := make(map[string][]interface{})
	for _, mapping := range mappings {
		for _, paramName := range mapping.ParameterNames {
			result[strings.ToLower(paramName)] = mapping.Rules
		}
	}
	return result
}

func resolveValidationRules(rulesInterface []interface{}, ruleDefinitions map[string]*ValidationRuleDefinition) []ValidationRule {
	result := make([]ValidationRule, 0)

	for _, ruleInterface := range rulesInterface {
		switch rule := ruleInterface.(type) {
		case string:
			// Simple rule name
			if def, found := ruleDefinitions[strings.ToLower(rule)]; found {
				result = append(result, ValidationRule{
					Name:       rule,
					Definition: def,
					Parameters: make(map[string]interface{}),
				})
			} else {
				common2.LogWarn("Validation rule '%s' not found in definitions", rule)
			}

		case map[string]interface{}:
			// Rule with parameters
			if name, ok := rule["Name"].(string); ok {
				if def, found := ruleDefinitions[strings.ToLower(name)]; found {
					params := make(map[string]interface{})
					for k, v := range rule {
						if k != "Name" {
							params[k] = v
						}
					}
					result = append(result, ValidationRule{
						Name:       name,
						Definition: def,
						Parameters: params,
					})
				} else {
					common2.LogWarn("Validation rule '%s' not found in definitions", name)
				}
			}
		}
	}

	return result
}
