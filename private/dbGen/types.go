package dbGen

import "github.com/keenmate/db-gen/private/version"

// Types used in template
type Property struct {
	DbColumnName   string
	DbColumnType   string
	PropertyName   string
	PropertyType   string
	Position       int
	MapperFunction string
	Nullable       bool // This can be unreliable
	Optional       bool // only used in Params
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
