# db-gen

Successor to ecto-gen

## Known Limitations

- generating code for functions that return one value acts like the function return void

## Configuration

All configuration is stored in file specified with `--config` flag or in `./db-gen.json`
Enable verbose logging with `--verbose` flag (can)

### Configuration overview

- **ConnectionString (string)**:
    - Defines the PostgreSQL database connection string.
    - For example `postgresql://usernmae:password@localhost:5432/database_name`
- **OutputFolder (string)**:
    - Specifies the folder where generated code files will be saved.
    - It can be relative to current working directory
- **OutputNamespace (string)**:
    - Sets the namespace for the generated code.
- **GenerateModels (boolean)**:
    - If **True** Generates models
- **GenerateProcessors (boolean)**:
    - If **True** Generates processors
- **SkipModelGenForVoidReturns (boolean)**:
    - If **True** doesn't generate models for functions without return type
- **ClearOutputFolder (boolean)**:
    - If **True** deletes content of output folder before generating new files
    - If this is false, program will show you which files have changed
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

processor and model templates have `Function` struct availible as `.` argument
dbcontext has `DbContextData` struct as `.` argument

```go
package main

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
	DbFunctionName     string
	HasReturn          bool
	IsProcedure        bool
	Parameters         []Property
	ReturnProperties   []Property
}

type DbContextData struct {
	Functions []Function
}

```