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

type Function struct {
	FunctionName       string
	DbFullFunctionName string
	ModelName          string
	ProcessorName      string
	Schema             string
	HasReturn          bool
	IsProcedure        bool
	Parameters         []Property
	ReturnProperties   []Property
}

type DbContextData struct {
	Functions []Function
}

type Command string

const (
	Init Command = "init"
	Gen          = "gen"
)

type Config struct {
	Command                    Command
	ConnectionString           string         `json:"ConnectionString"`
	OutputFolder               string         `json:"OutputFolder,omitempty"`
	OutputNamespace            string         `json:"OutputNamespace,omitempty"`
	GenerateModels             bool           `json:"GenerateModels,omitempty"`
	GenerateProcessors         bool           `json:"GenerateProcessors,omitempty"`
	SkipModelGenForVoidReturns bool           `json:"SkipModelGenForVoidReturns,omitempty"`
	DbContextTemplate          string         `json:"DbContextTemplate,omitempty"`
	ModelTemplate              string         `json:"ModelTemplate,omitempty"`
	ProcessorTemplate          string         `json:"ProcessorTemplate,omitempty"`
	Verbose                    bool           `json:"Verbose,omitempty"`
	ClearOutputFolder          bool           `json:"ClearOutputFolder,omitempty"`
	Generate                   []SchemaConfig `json:"Generate,omitempty"`
}

type SchemaConfig struct {
	Schema           string   `json:"Schema,omitempty"`
	AllFunctions     bool     `json:"AllFunctions,omitempty"`
	Functions        []string `json:"Functions,omitempty"`
	IgnoredFunctions []string `json:"IgnoredFunctions,omitempty"`
}
