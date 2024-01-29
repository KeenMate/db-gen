# db-gen

Successor to ecto-gen

## Known Limitations

- generating code for functions that return one value acts like the function return void

## Configuration

All configuration is stored in file specified with `--config` flag.
If `--config` flag is not set it will try following default locations

- `./db-gen.json`
- `./db-gen/db-gen.json`
- `./db-gen/config.json`

Enable debug logging with `--debug` flag

ConnectionString can be also set with `--connectionString "postgresql://usernmae:password@host:port/database_name"`

### Local configuration

For some secret or user-specific configuration, you can use local config.
Db-gen looks for file with prefix `local.` or `.local.` to loaded configuration.
The file can override any settings specified in the configuration file

For example, if you load configuration from `./db-gen/config.json` it will check fi `./db-gen/local.config.json` exists
and if it does, it will load it.

### Configuration overview

- **ConnectionString (string)**:
	- Defines the PostgreSQL database connection string.
	- For example `postgresql://usernmae:password@localhost:5432/database_name`
- **OutputFolder (string)**:
	- Specifies the folder where generated code files will be saved.
	- It can be relative to the current working directory
- **ProcessorsFolderName (string)**
	- folder name in output folder where processors will be generated
	- folder will be created if missing
- **ModelsFolderName (string)**
	- folder name in output folder where models will be generated
	- folder will be created if missing
- **GenerateModels (boolean)**:
	- If **True** Generates models
- **GenerateProcessors (boolean)**:
	- If **True** Generates processors
- **GenerateProcessorsForVoidReturns (boolean)**:
	- If **True** it generates processor even for functions that don't return anything
- **ClearOutputFolder (boolean)**:
	- If **True** deletes content of output folder before generating new files
- **DbContextTemplate (string)**:
	- Path to the template file for generating the dbContext file.
- **ModelTemplate (string)**:
	- Path to the template file for generating model file.
- **ProcessorTemplate (string)**:
	- Path to the template file for generating processor file.
- **GeneratedFileExtension (string)**:
	- Defines the file extension for generated files.
- **Generate**:
	- **Schema (string)**:
		- Specifies the database schema name.
	- **AllFunctions (boolean)**:
		- If true generated all functions except
	- **IgnoredFunctions (array of strings)**:
		- Functions to be ignored when generating code in the schema.
	- **Functions (array of strings)**:
		- Functions to be explicitly included when generating code in the schema.
- **Mappings**
	- **DatabaseTypes (array of strings)**:
		- If one database type has multiple mappings, last will be used
	- **MappedType (string)**:
		- Can be used in template
	- **MappingFunction (string)**:
		- Can be used in template

## Templates

Templates have there information available in `.`

```go
// DbContextTemplate
type DbContextData struct {
Config    *Config
Functions []Routine
BuildInfo *BuildInformation
}

//ProcessorTemplate
type ProcessorTemplateData struct {
Config    *Config
Routine   Routine
BuildInfo *BuildInformation
}

//ModelTemplate
type ModelTemplateData struct {
Config    *Config
Routine   Routine
BuildInfo *BuildInformation
}

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

```

### Case

By default, all fields use camel case.
You should use `pascalCased`/`camelCased`/`snakeCased` to change the case.
For example:

```gotemplate
{{pascalCased $func.FunctionName}}
```
