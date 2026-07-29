# Configuration reference

Every db-gen setting lives in a JSON config file (see [usage → configuration file resolution](./usage.md#configuration-file-resolution) for how it's located and how local overrides merge). This page is the source of truth for each key, organized by area.

> Source of truth for the schema: `Config` and related structs in `private/dbGen/config.go`. Boolean keys default to `false` unless noted.

## Contents

- [Connection & output](#connection--output)
- [What to generate](#what-to-generate)
- [Templates](#templates)
- [Offline generation](#offline-generation)
- [Context parameters](#context-parameters)
- [Parameter security](#parameter-security)
- [Additional generators](#additional-generators)
- [Copy targets](#copy-targets)
- [Generate — schema selection](#generate--schema-selection)
- [Mappings — type mapping](#mappings--type-mapping)
- [Validation](#validation)

## Connection & output

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ConnectionString` | string | — | PostgreSQL connection string, e.g. `postgresql://user:pass@localhost:5432/db`. Can be overridden with `--connectionString`. Best kept in a local override file. |
| `OutputFolder` | string | — | Folder for generated files. Relative to the config file's directory. Created if missing. |
| `ModelsFolderName` | string | `"models"` | Subfolder (under `OutputFolder`) for models. |
| `ProcessorsFolderName` | string | `"processors"` | Subfolder for processors. |
| `GeneratedFileExtension` | string | — | Extension for generated files, e.g. `.cs`, `.go`, `.ts`. |
| `GeneratedFileCase` | string | — | Filename case. One of `"snakecase"`, `"camelcase"`, `"pascalcase"`. Required (validated). |
| `ClearOutputFolder` | bool | `false` | Delete the output folder's contents before generating. Mutually exclusive with `RemoveOrphanedFiles`. |
| `RemoveOrphanedFiles` | bool | `false` | Remove previously generated files whose database function no longer exists. Tracks all output folders (main + additional generators). Only works when `ClearOutputFolder` is `false`. |
| `Debug` | bool | `false` | Enable debug logging and debug-file output. Usually set via the `--debug`/`-d` flag rather than in the config file. |

## What to generate

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `GenerateModels` | bool | `false` | Generate model types for routine return shapes. |
| `GenerateProcessors` | bool | `false` | Generate processors (mappers from db reader → model). |
| `GenerateProcessorsForVoidReturns` | bool | `false` | Also generate a processor for functions that return nothing. |
| `Generate` | array | — | Schema/function selection — see [below](#generate-schema-selection). |
| `Mappings` | array | — | Type mappings — see [below](#mappings-type-mapping). |

## Templates

| Key | Type | Description |
|-----|------|-------------|
| `DbContextTemplate` | string | Template that generates the database-call code. Path relative to config dir. |
| `ModelTemplate` | string | Template for model files. |
| `ProcessorTemplate` | string | Template for processor files. |

See [templating](./templating.md) for the data each template receives.

## Offline generation

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `RoutinesFile` | string | `"./db-gen-routines.json"` | Path to the JSON routine-definitions file (written by the `routines` command, read when offline). |
| `UseRoutinesFile` | bool | `false` | Read routines from `RoutinesFile` instead of connecting to the database. |

## Context parameters

Inject parameters like `_user_id` / `created_by` from a context object instead of passing them at every call site. Full guide: [context-mapping.md](./context-mapping.md).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `UseUserContext` | bool | `false` | Enable context-parameter mapping. |
| `UserContextParameterName` | string | `"ctx"` | Name of the context parameter in generated code. |
| `UserContextType` | string | `"UserContext"` | Type of the context parameter. |
| `ContextParameterMappings` | array | — | Each entry maps db parameter names to a context path. See below. |

`ContextParameterMappings[]`:

| Field | Type | Description |
|-------|------|-------------|
| `ParameterNames` | string[] | Database parameter names to treat as context (e.g. `["_user_id", "_userid"]`). |
| `ContextPath` | string | Property path on the context object, e.g. `"User.UserId"`. |

## Parameter security

Assign a logging-sensitivity level (`none` / `secure` / `strict` / `omit`) to parameters so templates can generate secure logging. db-gen resolves the level onto each parameter and exposes it as `Property.SecurityLevel`; it never renders masking itself — see [templating → parameter security levels](./templating.md#parameter-security-levels) for the level semantics and a template example.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `DefaultParameterSecurityLevel` | string | `"secure"` | Level used for any parameter not matched by the rules below. One of `none`/`secure`/`strict`/`omit` (validated, case-insensitive). |
| `ParameterSecurityMappings` | array | — | Global by-name assignments. Same shape as `ContextParameterMappings`. |

`ParameterSecurityMappings[]`:

| Field | Type | Description |
|-------|------|-------------|
| `ParameterNames` | string[] | Database parameter names this level applies to (e.g. `["password", "secret"]`). Case-insensitive. |
| `SecurityLevel` | string | One of `none`/`secure`/`strict`/`omit`. |

Per-function overrides take precedence over the global mappings: set `SecurityLevel` inside a function's [`Parameters` override object](./templating.md#per-routine-overrides). Resolution precedence is **per-function override → global mapping → `DefaultParameterSecurityLevel`**.

```json
{
  "DefaultParameterSecurityLevel": "secure",
  "ParameterSecurityMappings": [
    { "ParameterNames": ["password", "secret", "token"], "SecurityLevel": "strict" },
    { "ParameterNames": ["internal_note"], "SecurityLevel": "omit" }
  ],
  "Generate": [{
    "Schema": "public",
    "Functions": {
      "create_user": { "Parameters": { "password": { "SecurityLevel": "omit" } } }
    }
  }]
}
```

## Additional generators

Generate outputs beyond the three core templates (TypeScript types, providers, WebSocket endpoints, …). Full guide: [context-mapping.md → Additional Generators](./context-mapping.md).

`AdditionalGenerators[]`:

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Generator name (for logging). |
| `Enabled` | bool | Whether to run it. |
| `Template` | string | Template path. |
| `OutputFolder` | string | Output folder. |
| `FileName` | string | File name (for `single-file` generation). |
| `FileExtension` | string | File extension (for `per-routine` generation). |
| `FileCase` | string | Case style: `snakecase` / `camelcase` / `pascalcase`. |
| `GenerationType` | string | `"single-file"` or `"per-routine"`. Defaults to `single-file` if `FileName` is set, else `per-routine`. |
| `CleanOutputFolder` | bool | Delete the output folder before generating. |

## Copy targets

Generate bulk-`COPY`-into-table code from a table's columns. Full guide: [copy-targets.md](./copy-targets.md).

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `GenerateCopyTargets` | bool | `false` | Enable copy-target generation. |
| `CopyTargetTemplate` | string | — | Template for copy-target files. |
| `CopyTargetsFolderName` | string | `"copy"` | Subfolder under `OutputFolder` for copy-target files. |
| `CopyTargets` | array | — | Tables to generate for. See below. |

`CopyTargets[]`:

| Field | Type | Description |
|-------|------|-------------|
| `Schema` | string | Table schema. |
| `Table` | string | Table name. |
| `MappedName` | string | Override the generated struct/file name. |
| `Format` | string | Wire-format hint for the template: `"csv"` (default), `"text"`, or `"binary"`. |
| `NullString` | string | NULL sentinel for text/CSV formats. |

## Generate — schema selection

`Generate` is an array of schema configs:

| Field | Type | Description |
|-------|------|-------------|
| `Schema` | string | Database schema name. |
| `AllFunctions` | bool | Generate all functions in the schema except ones explicitly set to `false` in `Functions`. |
| `Functions` | object | Map of function name → `true`/`false` or a per-function override object. A key may be just the name (`my_func`) or name-with-params (`my_func(text,int)`). See [templating → per-routine overrides](./templating.md#per-routine-overrides). |

## Mappings — type mapping

`Mappings` is an array; each entry maps one or more PostgreSQL types to a target-language type. If a type appears in multiple entries, the last wins. See [templating → type mapping](./templating.md#three-tier-type-mapping) for resolution rules.

| Field | Type | Description |
|-------|------|-------------|
| `DatabaseTypes` | string[] | PostgreSQL `udt_name`s this entry covers (e.g. `["int4"]`). Use `"*"` as a catch-all fallback. |
| `MappedType` | string | Base type for non-nullable, non-optional cases. |
| `MappingFunction` | string | Function used to read the value from the db reader. |
| `NullableReturnType` | string | Type for nullable return values / model properties (e.g. `"int?"`). |
| `NullableParameterType` | string | Type for nullable parameters with no DEFAULT (e.g. `"int?"`). |
| `OptionalParameterType` | string | Type for optional parameters with a DEFAULT (e.g. `"Optional<int>"`). |

## Validation

db-gen can attach reusable validation rules to parameters. This is a larger subsystem documented separately in [validation.md](./validation.md); the corresponding `Validation` config object is described there.
