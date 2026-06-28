# db-gen

[![Release](https://img.shields.io/github/v/release/KeenMate/db-gen?sort=semver)](https://github.com/KeenMate/db-gen/releases)

> A language-agnostic CLI that generates database-access code from PostgreSQL stored functions and procedures — consistent, customizable, and fully offline-capable.

## What is it

`db-gen` connects to a PostgreSQL database, reads the metadata of your stored functions and procedures, and generates the boilerplate code that calls them — in whatever language and style your templates define. It is **not** an ORM: you keep writing SQL functions, and db-gen writes the typed call/mapping code around them.

Everything db-gen needs — configuration and templates — lives in your repository. The tool is a single self-contained executable with no runtime dependencies, no internet requirement, and no framework lock-in. The same binary generates the same code today and in five years.

**Headline features:**

- **Language-agnostic** — output is driven entirely by Go templates you control; C#, Go, Elixir, TypeScript, anything.
- **DbContext / Models / Processors** core generation, plus an extensible **Additional Generators** framework for any other output (TypeScript types, providers, WebSocket endpoints, …).
- **Three-tier type mapping** — distinct types for base, nullable returns, nullable params, and optional (DEFAULT) params.
- **Context parameter injection** — map `_user_id`, `created_by`, `tenant_id`, … to a context object instead of threading them through call sites.
- **Copy targets** — generate bulk-`COPY`-into-staging-table code from a table's columns, language-agnostically.
- **Schema change detection** — see what changed in your routines and copy-target tables since the last generation.
- **Offline generation** — export routine metadata to JSON and generate without a database connection.

## What's New

### Unreleased

- **Copy targets — generic bulk-`COPY` code generation.** Declare a staging table and db-gen emits the metadata (ordered columns, type, nullability, context/data split, format hint) a per-language template needs to write a correct `COPY ... FROM STDIN`. No language specifics live in the tool. Schema drift on those tables is now reported in `generate` and `database-changes` output, so a changed staging column is no longer silent. See [docs/copy-targets.md](./docs/copy-targets.md).
- **Test framework.** A three-layer suite (pure unit, DB-backed integration, golden-file e2e) plus `make test` / `test-unit` / `test-integration` / `test-force` / `test-update-golden`. See [test/README.md](./test/README.md).

### v0.6.1

- **Field-name generation preserves underscores before numbers** — `country_iso_2` no longer collapses to `country_iso2`. New `normalizeStr` template function uses normalized db column names directly instead of round-tripping through PascalCase.
- **Unnamed parameters** (`$1`, `$2`) no longer crash routine loading — `parameter_name` is coalesced to an empty string.

### v0.6.0

- **Context Parameter Mapping** — inject context params from a UserContext object ([docs/context-mapping.md](./docs/context-mapping.md)).
- **Additional Generators framework** — single-file or per-routine custom outputs.
- **Three-tier type mapping** — `NullableReturnType`, `NullableParameterType`, `OptionalParameterType`.
- **RemoveOrphanedFiles** — delete generated files when their database function disappears.

See [CHANGELOG.md](./CHANGELOG.md) for the full history.

## Docs

- 📘 [Usage & CLI reference](./docs/usage.md) — commands, flags, config resolution, offline workflow, change detection.
- ⚙️ [Configuration reference](./docs/configuration.md) — every config key, defaults, and what it does.
- 🧩 [Templating](./docs/templating.md) — template data model, functions, per-routine overrides, type mapping.
- 📦 [Copy targets](./docs/copy-targets.md) — bulk-`COPY` metadata contract and example templates.
- 🔐 [Context mapping & additional generators](./docs/context-mapping.md)
- ✅ [Validation](./docs/validation.md)
- 📚 [Examples / cookbook](./docs/examples.md)
- 🗂️ [Feature inventory](./FEATURES.md) — every feature, its config key, default, and test coverage.

## Install

Download the latest `db-gen-win.exe` / `db-gen-linux` from the [Releases page](https://github.com/KeenMate/db-gen/releases) and commit it next to your code. Keeping the binary in the repo is intentional — in five years it will still be there and still generate identical code.

Or build from source (Go 1.21+):

```bash
go build -o db-gen-win.exe .       # Windows
go build -o db-gen-linux .         # Linux
```

## Quick start

The fastest way to see it work is the bundled `test/` setup:

1. Create the test database from the migrations / SQL in `test/`.
2. Copy `test/db-gen.json` and set your connection string (a `local.db-gen.json` override keeps secrets out of the committed config).
3. Run generation:

```bash
db-gen generate --config test/db-gen.json
```

A minimal config:

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
    { "DatabaseTypes": ["text"], "MappedType": "string", "MappingFunction": "GetString" }
  ]
}
```

See [docs/usage.md](./docs/usage.md) for the full command surface and [docs/examples.md](./docs/examples.md) for richer recipes.

## CLI commands

| Command | What it does |
|---------|--------------|
| `generate` | Connect, read routines, and generate code. Reports schema changes since last run. |
| `routines [out]` | Export routine metadata to JSON for offline generation. |
| `database-changes` | Show what changed in the database since the last generation. |
| `completion [shell]` | Emit a shell-completion script (`bash`, `zsh`, `fish`, `powershell`). |
| `version` | Print version / build information. |

Global flags: `--config`/`-s` (config file path), `--connectionString`/`-c`, `--debug`/`-d`. See [docs/usage.md](./docs/usage.md).

## Why db-gen

- **Consistency over years** — a small versioned binary outlives the frameworks and SaaS tools it replaces. We lost projects to LLBLGen + .NET framework churn; db-gen avoids that by depending on nothing external.
- **In-house templates & config** — generation is fully under your control and lives in the repo; nothing depends on an internet service or installed tool.
- **Customization** — any language, db driver, or logger; just edit the template. You can even swap templates per environment (e.g. strip dev-only logging) in CI/CD.
- **Offline** — air-gapped, security-sensitive, or 2nd-base-camp-on-K2 environments are fully supported. Export routines to JSON once, generate forever without a connection.

## Architecture

Standard Go CLI: `main.go` → `cmd/` (Cobra commands) → `private/dbGen` (generation core), `private/database` (PostgreSQL access), `private/helpers`, `private/version`. The generation flow and a full diagram are in [docs/usage.md → How it works](./docs/usage.md#how-it-works).

## Development

```bash
go build -o db-gen-win.exe .     # build

make test                # all tests (DB-backed ones skip if no DB)
make test-unit           # fast, pure-Go unit tests only
make test-integration    # DB-backed integration + e2e
make test-force          # all tests, ignoring the Go test cache
make test-update-golden  # regenerate e2e golden files after intentional changes
```

See [test/README.md](./test/README.md) for the test framework layout.

## License

Proprietary to [Keenmate](https://github.com/keenmate) unless a `LICENSE` file in this repository states otherwise.

## Credits

Created by [Keenmate](https://github.com/keenmate). The field-naming behavior is kept consistent with the original `ecto_gen` tool.
