# Usage & CLI reference

This document covers how to run db-gen: the commands, their flags, how configuration is located, the offline workflow, and schema-change detection.

## Commands

```
db-gen <command> [flags]
```

| Command | Description |
|---------|-------------|
| `generate` | Connect to the database (or load a routines file), read routine metadata, apply filters/mappings, and generate code. Prints a summary of schema changes since the last generation. |
| `routines [out]` | Read routines from the database and write their definitions to a JSON file (default `./db-gen-routines.json`, or the `out` argument). Use this to enable offline generation later. |
| `database-changes` | Compare the current database schema against the last generation and print what changed — without generating anything. |
| `completion [bash\|zsh\|fish\|powershell]` | Print a shell-completion script. |
| `version` | Print version and build information. |
| `help [command]` | Print help for a command. |

## Flags

These flags are available on the data commands (`generate`, `routines`, `database-changes`):

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-s` | Path to the configuration file. If omitted, default locations are tried (see below). |
| `--connectionString` | `-c` | PostgreSQL connection string, e.g. `postgresql://user:pass@host:5432/db`. Overrides the value in config. |
| `--debug` | `-d` | Print debug logs and write debug files. |

## Configuration file resolution

If `--config` is not given, db-gen tries these paths in order and uses the first that exists:

1. `./db-gen.json`
2. `./db-gen/db-gen.json`
3. `./db-gen/config.json`

All relative paths inside the config (templates, output folder, routines file) are resolved **relative to the config file's directory**, not the current working directory.

### Local overrides

For secrets or per-developer settings, db-gen merges a local override file (if present) on top of the main config. Given a config at `testing/db-gen.json`, it looks for, in order:

- `testing/local.db-gen.json`
- `testing/.local.db-gen.json`
- `testing/db-gen.local`
- `testing/db-gen.json.local`

The first match is merged over the main config. The local file is optional and is the recommended place to keep your `ConnectionString` out of version control.

See the [configuration reference](./configuration.md) for every key.

## Offline generation

When you have no database connection available (air-gapped, CI, etc.), generate in two phases:

1. **Online, once** — export routine metadata:

   ```bash
   db-gen routines --config db-gen.json
   ```

   This writes `RoutinesFile` (default `./db-gen-routines.json`).

2. **Offline, any time** — set `"UseRoutinesFile": true` in config (or rely on it), then:

   ```bash
   db-gen generate --config db-gen.json
   ```

   db-gen reads routine metadata from the JSON file instead of connecting.

> Note: copy targets read column metadata directly from `information_schema` and currently require a database connection.

## Change detection

db-gen records the routine and copy-target metadata it generated against (in a generation-information file in the output folder). On the next run it diffs the live schema against that record.

- `generate` prints the changes and then regenerates.
- `database-changes` prints the changes **only** — useful in CI to detect drift without touching files.

Reported changes include, per routine: created / deleted functions, renamed parameters, parameter type and nullability changes, and added/removed parameters. For copy-target tables: created / deleted tables and per-column add/remove, type, and nullability changes.

This is what closes the gap where a changed staging-table column used to be silent — copy-target tables are tracked alongside routines.

## How it works

```mermaid
graph TD
    A[CLI Commands] --> B{Command Type}
    B -->|generate| C[Load Configuration]
    B -->|routines| D[Export Routines]
    B -->|database-changes| E[Detect Changes]

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
    N -->|Copy targets| CT[Apply Copy Target Template]
    N -->|Additional| AG[Apply Additional Generators]

    O --> U[Output Folder]
    P --> U
    Q --> U
    CT --> U
    AG --> U

    D --> V[Save Routines JSON]
    E --> W[Compare Schema Changes]

    subgraph "Configuration"
        X[db-gen.json]
        Y[local override]
        Z[Templates/*.gotmpl]
    end

    C --> X
    C --> Y
    O --> Z
    P --> Z
    Q --> Z
    CT --> Z
```

1. Load configuration (`db-gen.json` + optional local override).
2. Connect to PostgreSQL (or load the routines JSON file).
3. Query routine/column metadata from `information_schema`.
4. Apply filters (schema/function selection) and type mappings.
5. Render templates → write files to the output folder.

## See also

- [Configuration reference](./configuration.md)
- [Templating](./templating.md)
- [Copy targets](./copy-targets.md)
- [Examples / cookbook](./examples.md)
