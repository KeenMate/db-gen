# Examples & cookbook

Recipes for common db-gen setups. The bundled `test/` folder has runnable configs and templates you can copy from:

| Config | What it shows |
|--------|---------------|
| [`test/db-gen.json`](../test/db-gen.json) | Baseline generation (DbContext + models + processors) |
| [`test/db-gen-with-context.json`](../test/db-gen-with-context.json) | Context-parameter injection |
| [`test/db-gen-with-validation.json`](../test/db-gen-with-validation.json) | Parameter validation rules |
| [`test/db-gen-copy.json`](../test/db-gen-copy.json) | Copy targets (bulk `COPY`) |
| [`test/templates/`](../test/templates) | Example templates: `dbcontext`, `model`, `processor`, `typescript`, `provider`, `copy-pgx`, `copy-npgsql`, validation |

Run any of them:

```bash
db-gen generate --config test/db-gen.json
```

## Minimal config

Generate just the DbContext call code, no models/processors:

```json
{
  "ConnectionString": "postgresql://user:pass@localhost:5432/mydb",
  "OutputFolder": "./generated",
  "DbContextTemplate": "./templates/dbcontext.gotmpl",
  "GeneratedFileExtension": ".cs",
  "GeneratedFileCase": "pascalcase",
  "Generate": [{ "Schema": "public", "AllFunctions": true }],
  "Mappings": [
    { "DatabaseTypes": ["int4"], "MappedType": "int", "MappingFunction": "GetInt32" },
    { "DatabaseTypes": ["text"], "MappedType": "string", "MappingFunction": "GetString" },
    { "DatabaseTypes": ["*"], "MappedType": "object", "MappingFunction": "GetValue" }
  ]
}
```

The `"*"` mapping is a catch-all so an unmapped type doesn't stop generation.

## Models + processors

Add model and processor generation with nullable-aware types:

```json
{
  "GenerateModels": true,
  "GenerateProcessors": true,
  "ModelTemplate": "./templates/model.gotmpl",
  "ProcessorTemplate": "./templates/processor.gotmpl",
  "ModelsFolderName": "models",
  "ProcessorsFolderName": "processors",
  "Mappings": [
    {
      "DatabaseTypes": ["int4"],
      "MappedType": "int",
      "MappingFunction": "GetInt32",
      "NullableReturnType": "int?",
      "NullableParameterType": "int?",
      "OptionalParameterType": "Optional<int>"
    }
  ]
}
```

See [templating → three-tier type mapping](./templating.md#three-tier-type-mapping) for how those four types are selected.

## Selecting and overriding functions

Generate everything in a schema except a couple, and rename one:

```json
{
  "Generate": [
    {
      "Schema": "public",
      "AllFunctions": true,
      "Functions": {
        "internal_helper": false,
        "get_user_by_id": { "MappedName": "GetUser" }
      }
    }
  ]
}
```

Override a single return column's type/nullability:

```json
{
  "Functions": {
    "get_report": {
      "Model": {
        "total": { "MappedType": "decimal", "IsNullable": true }
      }
    }
  }
}
```

See [templating → per-routine overrides](./templating.md#per-routine-overrides).

## Context parameters

Stop threading `_user_id` / `created_by` through every call — map them to a context object:

```json
{
  "UseUserContext": true,
  "UserContextParameterName": "ctx",
  "UserContextType": "UserContext",
  "ContextParameterMappings": [
    { "ParameterNames": ["_user_id", "_userid"], "ContextPath": "User.UserId" },
    { "ParameterNames": ["created_by"], "ContextPath": "User.Username" }
  ]
}
```

In templates, `Routine.ContextParameters` and `Routine.RegularParameters` give you the two groups. Full guide: [context-mapping.md](./context-mapping.md).

## Additional generators (e.g. TypeScript types)

Generate a TypeScript model per routine alongside the core output:

```json
{
  "AdditionalGenerators": [
    {
      "Name": "TypeScript models",
      "Enabled": true,
      "Template": "./templates/typescript.gotmpl",
      "OutputFolder": "./generated/ts",
      "FileExtension": ".ts",
      "FileCase": "camelcase",
      "GenerationType": "per-routine",
      "CleanOutputFolder": true
    }
  ]
}
```

For a single aggregated file, set `FileName` instead of `FileExtension` and use `"GenerationType": "single-file"`.

## Copy targets (bulk COPY)

```json
{
  "GenerateCopyTargets": true,
  "CopyTargetTemplate": "./templates/copy-pgx.gotmpl",
  "CopyTargets": [
    { "Schema": "stage", "Table": "data_to_process", "Format": "text", "NullString": "" }
  ],
  "ContextParameterMappings": [
    { "ParameterNames": ["created_by"], "ContextPath": "ctx.CreatedBy" },
    { "ParameterNames": ["job_run_id"], "ContextPath": "ctx.JobRunId" }
  ]
}
```

Full guide: [copy-targets.md](./copy-targets.md).

## Offline generation

Export routine metadata once while connected, then generate without a database:

```bash
# online, once
db-gen routines --config db-gen.json      # writes db-gen-routines.json

# offline, any time (with "UseRoutinesFile": true)
db-gen generate --config db-gen.json
```

See [usage → offline generation](./usage.md#offline-generation).

## Detecting schema drift in CI

Fail a CI step when the database changed since the last generation, without writing files:

```bash
db-gen database-changes --config db-gen.json
```

See [usage → change detection](./usage.md#change-detection).
