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

### Core Generation Flow
1. Load configuration from `db-gen.json` (with optional local overrides)
2. Connect to PostgreSQL database via connection string
3. Query database for stored functions/procedures metadata
4. Apply filters and mappings based on configuration
5. Generate code files using Go templates

### Templates and Output
- Uses Go templates (.gotmpl files) for code generation
- Supports three template types: DbContext, Model, and Processor
- Templates receive structured data about database routines and configuration
- Output can be customized per language/framework via template modification

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

## Type Mapping System

The tool maps PostgreSQL types to target language types via:
1. Global mappings in configuration (DatabaseTypes → MappedType + MappingFunction)
2. Per-function overrides for custom handling
3. Template functions for case conversion (pascalCased, camelCased, snakeCased)

## Template Development

Templates use Go template syntax with access to:
- `Config` - Full configuration object
- `Functions` - Array of database routines (for DbContext template)
- `Routine` - Single routine data (for Model/Processor templates)
- `BuildInfo` - Version and build information

Each `Routine` includes function metadata, parameters, return properties, and naming information.