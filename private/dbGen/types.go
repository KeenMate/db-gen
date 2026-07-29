package dbGen

import "github.com/keenmate/db-gen/private/version"

// Types used in template
type Property struct {
	DbColumnName       string
	DbColumnType       string
	PropertyName       string
	PropertyType       string // The actual type to use (resolved based on nullable/optional)
	BaseType           string // Base non-nullable type
	NullableReturnType string // Type for nullable return values/model properties
	NullableParamType  string // Type for nullable parameters
	OptionalParamType  string // Type for optional parameters (with defaults)
	Position           int
	MapperFunction     string
	Nullable           bool // This can be unreliable
	Optional           bool // only used in Params
	IsContextParameter bool
	ContextPath        string
	ValidationRules    []ValidationRule
	SecurityLevel      string // Logging sensitivity: none | secure | strict | omit
}

type ValidationRule struct {
	Name       string
	Definition *ValidationRuleDefinition
	Parameters map[string]interface{}
}

type Routine struct {
	FunctionName       string
	DbFullFunctionName string
	ModelName          string
	ProcessorName      string
	Schema             string
	DbFunctionName     string
	HasReturn          bool
	IsProcedure        bool
	Parameters         []Property
	ReturnProperties   []Property
	UsesUserContext    bool
	ContextParameters  []Property
	RegularParameters  []Property
}

type DbContextData struct {
	Config    *Config
	Functions []Routine
	BuildInfo *version.BuildInformation
}

type ProcessorTemplateData struct {
	Config    *Config
	Routine   Routine
	BuildInfo *version.BuildInformation
}

type ModelTemplateData struct {
	Config    *Config
	Routine   Routine
	BuildInfo *version.BuildInformation
}
