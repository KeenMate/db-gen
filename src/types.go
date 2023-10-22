package dbGen

// Types used in template
type Property struct {
	DbColumnName   string
	DbColumnType   string
	PropertyName   string
	PropertyType   string
	Position       int
	MapperFunction string
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
	Functions []Routine
}

type Command string

const (
	Init Command = "init"
	Gen          = "gen"
)

type Config struct {
	Command                    Command
	PathBase                   string         //for now just using config folder
	ConnectionString           string         `json:"ConnectionString"`
	OutputFolder               string         `json:"OutputFolder,omitempty"`
	GenerateModels             bool           `json:"GenerateModels,omitempty"`
	GenerateProcessors         bool           `json:"GenerateProcessors,omitempty"`
	SkipModelGenForVoidReturns bool           `json:"SkipModelGenForVoidReturns,omitempty"`
	DbContextTemplate          string         `json:"DbContextTemplate,omitempty"`
	ModelTemplate              string         `json:"ModelTemplate,omitempty"`
	ProcessorTemplate          string         `json:"ProcessorTemplate,omitempty"`
	GeneratedFileExtension     string         `json:"GeneratedFileExtension,omitempty"`
	Verbose                    bool           `json:"Verbose,omitempty"`
	ClearOutputFolder          bool           `json:"ClearOutputFolder,omitempty"`
	Generate                   []SchemaConfig `json:"Generate,omitempty"`
	Mappings                   []Mapping      `json:"Mappings"`
}

type SchemaConfig struct {
	Schema           string   `json:"Schema,omitempty"`
	AllFunctions     bool     `json:"AllFunctions,omitempty"`
	Functions        []string `json:"Functions,omitempty"`
	IgnoredFunctions []string `json:"IgnoredFunctions,omitempty"`
}

// TODO in config.go validate that we dont have multiple mappings to one DatabaseType
type Mapping struct {
	DatabaseTypes   []string `json:"DatabaseTypes"`
	MappedType      string   `json:"MappedType"`
	MappingFunction string   `json:"MappingFunction"`
}
