# db-gen - Language agnostic database function calls code generator for Enterprise use

db-gen is a universal tool for generation of function calls to PostgreSQL database.

## Is this tool an ORM framework?

No, this tool is not an ORM framework in sense of C# Entity Framework, Elixir Ecto, PHP Doctrine and so on. In our experience, these full ORM tools are not
worth it, they are usually clumsy, generate inefficient SQL code and lead programmers to dead ends.
Typical example of inefficent database use is when you want a multi step processing of imported data. Instead of one bulk copy and a single database call to
process the data, you have to split your logic to multiple database calls. It's slower, more work and usually less safe.

That's why we use stored functions/procedures in PostgreSQL and this tool just generates code that calls these functions/procedures and retrieves the data.

## What issues this tool tries to address?

- consistency of generation over years
- in-house templates, in-house configuration
- customization based on your needs
- offline use

Don't let the "Enterprise use" discourage you, there is no reason for not to use this tool for your one function database.

## Consistency of generation over years

We all know what kind of world we live in. Tool that was available yesterday, won't be available tomorrow. Tool that was working with yesterday's framework,
won't be working with tomorrow's.

This is NOT sustainable in enterprise development.

It's not like every application is constantly being updated and pushed to the latest version of every package. We have application that are untouched for years
and years because of budget reasons. Why update them when they are running, right?
We used LLBLGen on several projects, but after just a few years we are unable to do that anymore, .NET framework was replace with another .NET framework and all
is lost.

That's why this tool goes a different way. It's a small executable package that can be easily stored to the repository with your code. It will generate the same
code today, tomorrow and in 5 years, and you won't have to search for it on internet.

## In-house templates, In-house configuration

All configuration, including templates used for code generation are part of the repository. Nothing depends on some service in internet, or a tool installation.
Everything is under your control, versioned, easily updateable.

## Customization based on your needs

Since everything is under your control, as mentioned above, you can use whatever language, database package, logger and so on. Just update the template and you
are done.

## Offline use

In Enterprise development, it is often the case that your internet connection is limited, or there is none, in case of security sensitive projects you might not
have internet at all. In case of digital nomads, you might be currently working in K2 2nd base camp. In all these cases you are covered, db-gen is
self-contained executable, it needs nothing else than configuration and templates.

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
Db-gen looks for file with prefix `local.` or `.local.` to loaded configuration
or with postfix `.local`.

So if we load config at `./testing/db-gen.json`
it will look at

- `testing\local.db-gen.json`
- `testing\.local.db-gen.json`
- `testing\db-gen.local`
- `testing\db-gen.json.local`

The loaded configuration will override the values set in normal config file.

The local config file is not required.

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
		- If true generated all functions except explicitly ignored by adding functions entry with false value
	- **Functions (object where values are bool or object)**:
		- Keys of object are function names, you can you only name, or name with parameters (`function(text,int)` =`function`)
		- If value is just bool, it only specifies if it should be generated
    - You can supply object and it will override global mappings see [Mapping](#Mapping-override-per-routines)
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




```

### Case

By default, all fields use camel case.
You should use `pascalCased`/`camelCased`/`snakeCased` to change the case.
For example:

```gotemplate
{{pascalCased $func.FunctionName}}
```


### Mapping override per routines

_TODO Improve this section_


You can specify custom mapping for each function, parameter and model by providing object to `Functions` properties

You can override: 

Name using `MappedName`, Processors and models name will be created by adding model/processor to this name

HasReturn using `DontRetrieveValues`, it can only be used to disable selection of function which has return, 
not other way around.

#### Model

Use `SelectOnlySpecified` to only select columns you explicitly specify in model by setting them to true,
or providing custom mapping

In `Model` provide object with where keys correspond to columns in database. 
If you set value to false, it will not select it. Setting value to true or providing object with mapping
will select it. 

In mapping object you can override `MappedName`, `IsNullable`, `MappedType` and `MappingFunction`.
If you only specify `MappedType` it will try to find mapping function in global mappings, stoping generation with error if it didnt.
Setting `MappingFunction` without `MappedType` will do nothing.

#### Parameters

It doesnt make sense to only use some parameter, so you can only change `MappedName`,`MappedType`, and `IsNUllable`
