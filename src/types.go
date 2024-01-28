package dbGen

// Types used in template
type Property struct {
	DbColumnName   string
	DbColumnType   string
	PropertyName   string
	PropertyType   string
	Position       int
	MapperFunction string
	Nullable       bool // This can be unreliable
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
}

type ProcessorTemplateData struct {
	Config  *Config
	Routine Routine
}

type ModelTemplateData struct {
	Config  *Config
	Routine Routine
}
