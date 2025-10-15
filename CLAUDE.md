# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

db-gen is a Go-based CLI tool for generating database access code from PostgreSQL stored functions and procedures. It's designed for enterprise use with emphasis on consistency, customization, and offline operation.

## Architecture

The codebase follows a standard Go CLI structure:

- `main.go` - Entry point that embeds version.txt and delegates to cmd package
- `cmd/` - Cobra CLI commands (root, generate, routines, databaseChanges, completion, version)
- `private/` - Internal packages:
  - `dbGen/` - Core generation logic, configuration, templates, and data processing
  - `database/` - PostgreSQL connection and query handling
  - `helpers/` - Utility functions (logging, filesystem, strings, timers)
  - `version/` - Build information handling

## Key Components

### Commands
- `generate` - Main command that connects to database and generates code files
- `routines` - Exports database function definitions to JSON for offline use
- `databaseChanges` - Detects schema changes since last generation
- `completion` - Generates shell completion scripts (bash, zsh, fish, powershell)

### Core Generation Flow
1. Load configuration from `db-gen.json` (with optional local overrides)
2. Connect to PostgreSQL database via connection string
3. Query database for stored functions/procedures metadata
4. Apply filters and mappings based on configuration
5. Generate code files using Go templates

### Templates and Output
- Uses Go templates (.gotmpl files) for code generation
- Supports three core template types: DbContext, Model, and Processor
- Additional generators support custom templates for other outputs (e.g., CommonProvider, TypeScript models)
- Templates receive structured data about database routines and configuration
- Output can be customized per language/framework via template modification
- Additional generators can generate either single-file or per-routine outputs

## Development Commands

### Build
```bash
go build -o db-gen-win.exe .       # Windows
go build -o db-gen-linux .         # Linux
```

### Run
```bash
go run main.go generate --config test/db-gen.json
go run main.go routines --config test/db-gen.json
go run main.go databaseChanges --config test/db-gen.json
```

### Test
Use the test database and configuration in the `test/` folder:
```bash
# Set up test database first using test/database/testing-db.sql
go run main.go generate --config test/db-gen.json
```

## Configuration

Configuration files follow the pattern:
- Primary: `db-gen.json`, `db-gen/db-gen.json`, or `db-gen/config.json`
- Local overrides: `local.db-gen.json`, `.local.db-gen.json`, etc.

Key configuration sections:
- **Connection**: PostgreSQL connection string
- **Output**: Folder paths, file extensions, template locations
- **Generation**: Schema selection, function filtering, type mappings
- **Templates**: Paths to Go template files for different output types
- **Context Parameter Mapping**: Automatic injection of user context parameters
- **Additional Generators**: Extensible system for custom code generation beyond DbContext/Models/Processors

## Type Mapping System

The tool uses a three-tier type mapping system:
1. **Base Type (`MappedType`)**: Used for non-nullable, non-optional cases
2. **Nullable Types**:
   - `NullableReturnType`: For nullable return values in models (e.g., `int?`)
   - `NullableParameterType`: For nullable parameters with no DEFAULT (e.g., `int?`)
3. **Optional Type (`OptionalParameterType`)**: For parameters with DEFAULT values (e.g., `Optional<int>`)

Type resolution happens in mapper.go:
- Parameters check `isOptional` first, then `isNullable`
- If optional and `optionalParameterType` is set → use it
- Else if nullable and `nullableParameterType` is set → use it
- Otherwise use base `mappedType`

Per-function overrides can customize:
- `MappedType`: Override the base type
- `IsNullable`: Override nullable detection
- `IsOptional`: Override optional detection
- Template functions available: `pascalCased`, `camelCased`, `snakeCased`, `trimPrefix`

## Configuration Options

Key boolean settings (all default to false):
- `GenerateModels`, `GenerateProcessors`, `GenerateProcessorsForVoidReturns`
- `ClearOutputFolder`, `RemoveOrphanedFiles`, `UseRoutinesFile`

Valid enum values:
- `GeneratedFileCase`: "snakecase", "camelcase", "pascalcase"

## Template Development

Templates use Go template syntax with access to:
- `Config` - Full configuration object
- `Functions` - Array of database routines (for DbContext template)
- `Routine` - Single routine data (for Model/Processor templates)
- `BuildInfo` - Version and build information

Each `Routine` includes function metadata, parameters, return properties, and naming information.

## Recent Updates (v0.6.0)

### New Features
- **Context Parameter Mapping**: Automatic injection of context parameters (_user_id, _created_by, _tenant_id, etc.) from UserContext
  - Configure via `UseUserContext`, `UserContextParameterName`, `UserContextType`
  - Define mappings in `ContextParameterMappings` array
  - Parameters are split into context vs regular in generated code

- **Additional Generators Framework**: Generate custom outputs beyond DbContext/Models/Processors
  - Support for single-file generation (e.g., CommonProvider)
  - Support for per-routine generation (e.g., TypeScript models)
  - Configurable via `AdditionalGenerators` array
  - Optional `CleanOutputFolder` to remove stale files

- **Three-tier Type Mapping**: Separate handling for nullable returns, nullable parameters, and optional parameters
  - `NullableReturnType`: For model properties (e.g., `int?`)
  - `NullableParameterType`: For nullable params without DEFAULT (e.g., `int?`)
  - `OptionalParameterType`: For params with DEFAULT (e.g., `Optional<int>`)

- **RemoveOrphanedFiles**: Automatically remove generated files when database functions are deleted
  - Set `RemoveOrphanedFiles: true` in config
  - Only works when `ClearOutputFolder: false` (mutually exclusive)
  - Tracks all output folders (main + AdditionalGenerators)
  - Uses MD5 hash comparison to detect orphaned files

### Bug Fixes
- Fixed per-function type override bug where nullable/optional types weren't loaded from global mappings
- Fixed nullable parameter handling to use `T?` instead of `Optional<T>` to prevent parameters from being filtered out
- Fixed double-wrapping issue in DbContext template
- Fixed jsonb parameter architectural boundary - keeps PostgreSQL types internal to DbContext
- Fixed change detection to track files from all output folders, preventing all TypeScript files showing as "Updated" when only one changed

### Template Data Updates
- Property struct now includes: `BaseType`, `NullableReturnType`, `NullableParamType`, `OptionalParamType`
- Property struct includes: `IsContextParameter`, `ContextPath` for context mapping
- Routine struct includes: `ContextParameters`, `RegularParameters`, `Parameters` (all params in order)

### Important Implementation Details
- **Nullable vs Optional distinction**:
  - Nullable (no DEFAULT, accepts NULL) → always pass to DB, can be null → `T?`
  - Optional (has DEFAULT) → can be omitted from DB call → `Optional<T>`
  - Templates check `$parameter.Optional` to decide `.ToObjectOptional()` vs `Optional.Some()`
- **Context parameter processing**: Done in `processContextParameters()` in mapper.go
- **Type resolution**: Happens in mapper.go lines 229-235, checks optional first, then nullable
- **Per-function overrides**: Must pass empty string (not hardcoded text) to `handleTypeMappingOverride()` to properly lookup global mappings