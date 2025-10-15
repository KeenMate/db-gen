# db-gen - Language agnostic database function calls code generator for Enterprise use

db-gen is a universal tool for generation of function calls to PostgreSQL database.

## Is this tool an ORM framework?

No, this tool is not an ORM framework in the sense of C# Entity Framework, Elixir Ecto, PHP Doctrine, and so on. In our experience, these full ORM tools are not
worth it; they are usually clumsy, generate inefficient SQL code, and lead programmers to dead ends.
A typical example of inefficient database use is when you want multi-step processing of imported data. Instead of one bulk copy and a single database call to
process the data, you have to split your logic into multiple database calls. It's slower, more work, and usually less safe.

That's why we use stored functions/procedures in PostgreSQL and this tool just generates code that calls these functions/procedures and retrieves the data.

## What issues this tool tries to address?

- Consistency of generation over years
- In-house templates and configuration
- Customization based on your needs
- Offline use

Don't let the "Enterprise use" discourage you, there is no reason for not to use this tool for your one function database.

## Consistency of generation over years

We all know what kind of world we live in. A tool that was available yesterday won't be available tomorrow. A tool that was working with yesterday's framework
won't be working with tomorrow's.

This is NOT sustainable in enterprise development.

It's not like every application is constantly being updated and pushed to the latest version of every package. We have applications that are untouched for years
and years because of budget reasons. Why update them when they are running, right?
We used LLBLGen on several projects, but after just a few years we are unable to do that anymore; .NET framework was replaced with another .NET framework and all
is lost.

That's why this tool goes a different way. It's a small executable package that can be easily stored in the repository with your code. It will generate the same
code today, tomorrow, and in 5 years, and you won't have to search for it on the internet.

## In-house templates and configuration

All configuration, including templates used for code generation, are part of the repository. Nothing depends on some service on the internet or a tool installation.
Everything is under your control, versioned, and easily updatable.

## Customization based on your needs

Since everything is under your control, as mentioned above, you can use whatever language, database package, logger, and so on. Just update the template and you
are done.

## Offline use

In enterprise development, it is often the case that your internet connection is limited, or there is none. In case of security-sensitive projects, you might not
have internet at all. In case of digital nomads, you might be currently working in K2's 2nd base camp. In all these cases you are covered; db-gen is a
self-contained executable that needs nothing else than configuration and templates.

## Architecture Overview

```mermaid
graph TD
    A[CLI Commands] --> B{Command Type}
    B -->|generate| C[Load Configuration]
    B -->|routines| D[Export Routines]
    B -->|databaseChanges| E[Detect Changes]

    C --> F{Data Source}
    F -->|Database| G[Connect to PostgreSQL]
    F -->|Offline| H[Load Routines File]

    G --> I[Query Database Schema]
    H --> J[Parse JSON Routines]
    I --> K[Extract Function Metadata]
    J --> K

    K --> L[Apply Filters & Mappings]
    L --> M[Generate Routine Objects]

    M --> N{Generation Type}
    N -->|DbContext| O[Apply DbContext Template]
    N -->|Models| P[Apply Model Template]
    N -->|Processors| Q[Apply Processor Template]

    O --> R[Generate DbContext File]
    P --> S[Generate Model Files]
    Q --> T[Generate Processor Files]

    R --> U[Output Folder]
    S --> U
    T --> U

    D --> V[Save Routines JSON]
    E --> W[Compare Schema Changes]

    subgraph "Configuration Files"
        X[db-gen.json]
        Y[local.db-gen.json]
        Z[Templates/*.gotmpl]
    end

    subgraph "Database"
        AA[PostgreSQL]
        AB[Stored Functions/Procedures]
    end

    subgraph "Output"
        AC[Generated Code Files]
        AD[DbContext]
        AE[Models]
        AF[Processors]
    end

    C --> X
    C --> Y
    O --> Z
    P --> Z
    Q --> Z
    G --> AA
    I --> AB
    U --> AC
    R --> AD
    S --> AE
    T --> AF
```

## How to use it

We usually put the downloaded `db-gen-win.exe`, `db-gen-linux`, or both directly into the repo. Yes, the repo gets bigger, but in five years, when you have to update your project, you won't have to look for it on the ever-forgetting internet.
Also, when it's part of the repo, you can run `db-gen generate` as part of your CD/CI and use different templates. Why would you do that? For example, to remove log messages that should be visible only in the development environment. This is what Erlang/Elixir does to speed up their code.

When you run `db-gen`, you are offered these main commands:
- `generate` - will run the generation of code
- `routines` - will generate a JSON file that contains definitions of all stored functions/procedures that you have defined in `db-gen.json`; this can later be used for offline generation
- `databaseChanges` - will detect changes in database schema since last generation
- `completion [shell]` - generate shell completion script
	- Valid shell values: `bash`, `zsh`, `fish`, `powershell`
- `help [command]` - will print out help for a specific command with additional details

## How to start

The easiest way is to:
- Take the content from the `test` folder
- Use the `test/database/testing-db.sql` script to create a test database
- Update the connection string to proper values in `test/local.db-gen.json`
- Download the latest `db-gen` release from the [Releases page](https://github.com/KeenMate/db-gen/releases)
- Run `./db-gen-win.exe generate` or `./db-gen-linux generate`
- Be properly amazed, shocked, and stunned!

## Configuration

All configuration is stored in the file specified with the `--config` flag.
If the `--config` flag is not set, it will try the following default locations (in order):

- `./db-gen.json`
- `./db-gen/db-gen.json`
- `./db-gen/config.json`

Enable debug logging with the `--debug` flag.

ConnectionString can also be set with `--connectionString "postgresql://username:password@host:port/database_name"`


### Local configuration

For some secret or user-specific configuration, you can use local config.
Db-gen looks for files with the following patterns relative to your main config file:

**Prefixes:** `local.` or `.local.`
**Postfixes:** `.local`

So if we load config at `./testing/db-gen.json`,
it will look for local overrides at:

- `testing/local.db-gen.json`
- `testing/.local.db-gen.json`
- `testing/db-gen.local`
- `testing/db-gen.json.local`

The loaded configuration will override the values set in normal config file.

The local config file is not required.

### Configuration overview

- **ConnectionString (string)**:
	- Defines the PostgreSQL database connection string.
	- For example: `postgresql://username:password@localhost:5432/database_name`
- **OutputFolder (string)**:
	- Specifies the folder where generated code files will be saved.
	- It can be relative to the current working directory
- **ProcessorsFolderName (string)**:
	- Folder name in output folder where processors will be generated (default: "processors")
	- Folder will be created if missing
- **ModelsFolderName (string)**:
	- Folder name in output folder where models will be generated (default: "models")
	- Folder will be created if missing
- **GenerateModels (boolean)**:
	- If `true`, generates models
	- Valid values: `true`, `false`
- **GenerateProcessors (boolean)**:
	- If `true`, generates processors
	- Valid values: `true`, `false`
- **GenerateProcessorsForVoidReturns (boolean)**:
	- If `true`, generates processor even for functions that don't return anything
	- Valid values: `true`, `false`
- **ClearOutputFolder (boolean)**:
	- If `true`, deletes content of output folder before generating new files
	- Valid values: `true`, `false`
- **RemoveOrphanedFiles (boolean)**:
	- If `true`, removes generated files when their corresponding database functions no longer exist
	- Only works when `ClearOutputFolder` is `false` (mutually exclusive)
	- Tracks all output folders (main + AdditionalGenerators)
	- Valid values: `true`, `false`
- **DbContextTemplate (string)**:
	- Path to the template file for generating the dbContext file
- **ModelTemplate (string)**:
	- Path to the template file for generating model files
- **ProcessorTemplate (string)**:
	- Path to the template file for generating processor files
- **GeneratedFileExtension (string)**:
	- Defines the file extension for generated files
- **GeneratedFileCase (string)**:
	- Defines the case style for generated files
	- Valid values: `"snakecase"`, `"camelcase"`, `"pascalcase"`
- **RoutinesFile (string)**:
	- Path to the JSON file containing routine definitions for offline generation (default: "./db-gen-routines.json")
- **UseRoutinesFile (boolean)**:
	- If `true`, uses the routines file for offline generation instead of connecting to database
	- Valid values: `true`, `false`
- **UseUserContext (boolean)**:
	- If `true`, enables automatic context parameter mapping
	- Valid values: `true`, `false`
- **UserContextParameterName (string)**:
	- Name of the user context parameter (e.g., `"ctx"`)
- **UserContextType (string)**:
	- Type of the user context parameter (e.g., `"UserContext"`)
- **ContextParameterMappings (array)**:
	- Maps database parameter names to user context properties
	- Each mapping has:
		- **ParameterNames (array of strings)**: Database parameter names to map (e.g., `["_user_id", "_userid"]`)
		- **ContextPath (string)**: Property path in context object (e.g., `"User.UserId"`)
- **AdditionalGenerators (array)**:
	- Defines additional code generation outputs beyond DbContext/Models/Processors
	- Each generator has:
		- **Name (string)**: Generator name for logging
		- **Enabled (boolean)**: Whether to run this generator
		- **Template (string)**: Path to template file
		- **OutputFolder (string)**: Where to save generated files
		- **FileName (string)**: File name for single-file generation
		- **FileExtension (string)**: File extension for per-routine generation
		- **FileCase (string)**: Case style for generated files (`"snakecase"`, `"camelcase"`, `"pascalcase"`)
		- **GenerationType (string)**: `"single-file"` or `"per-routine"`
		- **CleanOutputFolder (boolean)**: If `true`, deletes output folder before generation
- **Generate**:
	- **Schema (string)**:
		- Specifies the database schema name
	- **AllFunctions (boolean)**:
		- If `true`, generates all functions except those explicitly ignored by adding a functions entry with `false` value
		- Valid values: `true`, `false`
	- **Functions (object where values are bool or object)**:
		- Keys of the object are function names; you can use only the name, or name with parameters (`function(text,int)` = `function`)
		- If value is just bool (`true`/`false`), it only specifies if it should be generated
		- You can supply an object and it will override global mappings (see [Mapping override per routines](#mapping-override-per-routines))
- **Mappings**:
	- **DatabaseTypes (array of strings)**:
		- If one database type has multiple mappings, the last will be used
	- **MappedType (string)**:
		- Base type for non-nullable, non-optional cases
	- **MappingFunction (string)**:
		- Function used to retrieve value from database reader
	- **NullableReturnType (string)**:
		- Type to use for nullable return values in models (e.g., `"int?"`)
	- **NullableParameterType (string)**:
		- Type to use for nullable parameters (no DEFAULT, accepts NULL) (e.g., `"int?"`)
	- **OptionalParameterType (string)**:
		- Type to use for optional parameters (has DEFAULT value) (e.g., `"Optional<int>"`)

## Templates

Templates to use are defined in these properties of `db-gen.json`:
 - **DbContextTemplate** - this will generate database calls
 - **ModelTemplate** - this will generate models to represent data coming from the database
 - **ProcessorTemplate** - this will generate mappers mapping data from the database to models

Templates use database metadata in format:

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

Templates themselves are written in Go Templates and can be changed to your liking. You are in complete control.

### Case

By default, all fields use camel case.
You can use `pascalCased`, `camelCased`, or `snakeCased` to change the case.
For example:

```gotemplate
{{pascalCased $func.FunctionName}}
```

### Mapping override per routines

_TODO Improve this section_


You can specify custom mapping for each function, parameter, and model by providing an object to the `Functions` properties.

You can override:

**Name** using `MappedName` - processor and model names will be created by adding model/processor to this name

**HasReturn** using `DontRetrieveValues` - it can only be used to disable selection of functions that have return values,
not the other way around.

#### Model

Use `SelectOnlySpecified` to only select columns you explicitly specify in the model by setting them to true
or providing custom mapping.

In `Model`, provide an object where keys correspond to columns in the database.
If you set the value to false, it will not select it. Setting the value to true or providing an object with mapping
will select it.

In the mapping object, you can override `MappedName`, `IsNullable`, `MappedType`, and `MappingFunction`.
If you only specify `MappedType`, it will try to find the mapping function in global mappings, stopping generation with an error if it doesn't find one.
Setting `MappingFunction` without `MappedType` will do nothing.


#### Parameters

> **DISCLAIMER:** Needs clarification

It doesn't make sense to only use some parameters, so you can only change `MappedName`, `MappedType`, `IsNullable`, and `IsOptional`. This also means that you can't set a parameter value to boolean; you can only set it to an object with custom mapping.

### Overloaded function

> **DISCLAIMER:** Needs clarification

To prevent a LOT of issues with overloaded functions, you are forced to specify a mapped name for each function that has some overload.

The name has to be unique in the schema, but checking is not yet implemented, so be careful!
