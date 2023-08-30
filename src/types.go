package dbGen

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
	HasReturn          bool
	IsProcedure        bool
	Parameters         []Property
	ReturnProperties   []Property
}

type dbContextData struct {
	Functions []Function
}
