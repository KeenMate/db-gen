# db-gen

[![Release](https://img.shields.io/github/v/release/KeenMate/db-gen?sort=semver)](https://github.com/KeenMate/db-gen/releases)

> A language-agnostic CLI that generates database-access code from PostgreSQL stored functions and procedures — consistent, customizable, and fully offline-capable.

## Contents

- [What is it](#what-is-it) · [What's New](#whats-new) · [Install](#install) · [Quick start](#quick-start) · [CLI commands](#cli-commands) · [Why db-gen](#why-db-gen) · [Architecture](#architecture) · [Development](#development)
- **Reference docs:** [Usage & CLI](./docs/usage.md) · [**Configuration**](./docs/configuration.md) — every setting, grouped by area · [Templating](./docs/templating.md) · [Copy targets](./docs/copy-targets.md) · [Context mapping & generators](./docs/context-mapping.md) · [Validation](./docs/validation.md) · [Examples / cookbook](./docs/examples.md) · [Example templates](./examples) · [Feature inventory](./FEATURES.md)

See the annotated [Docs](#docs) index below for one-line descriptions of each.

## What is it

`db-gen` connects to a PostgreSQL database, reads the metadata of your stored functions and procedures, and generates the boilerplate code that calls them — in whatever language and style your templates define. It is **not** an ORM: you keep writing SQL functions, and db-gen writes the typed call/mapping code around them.

Everything db-gen needs — configuration and templates — lives in your repository. The tool is a single self-contained executable with no runtime dependencies, no internet requirement, and no framework lock-in. The same binary generates the same code today and in five years.

**Headline features:**

- **Language-agnostic** — output is driven entirely by Go templates you control; C#, Go, Elixir, TypeScript, anything.
- **DbContext / Models / Processors** core generation, plus an extensible **Additional Generators** framework for any other output (TypeScript types, providers, WebSocket endpoints, …).
- **Three-tier type mapping** — distinct types for base, nullable returns, nullable params, and optional (DEFAULT) params.
- **Context parameter injection** — map `_user_id`, `created_by`, `tenant_id`, … to a context object instead of threading them through call sites.
- **Copy targets** — generate bulk-`COPY`-into-staging-table code from a table's columns, language-agnostically.
- **Parameter security levels** — tag parameters `none`/`secure`/`strict`/`omit` so templates can generate secure logging (mask or omit sensitive values).
- **Schema change detection** — see what changed in your routines and copy-target tables since the last generation.
- **Offline generation** — export routine metadata to JSON and generate without a database connection.

## What's New

### v0.8.0

- **Template variables.** A new `TemplateVariables` config key holds a free-form string→string map that db-gen passes through to *every* template, read with the new `templateVar` function: `{{templateVar .Config.TemplateVariables "GeneratedNs"}}`. This lets one shared template set serve several projects that differ only in config — point each project's `db-gen.json` at its own namespaces and keep a single set of templates. Lookup is case-insensitive. See [docs/configuration.md → template variables](./docs/configuration.md#template-variables).
- **Parameter security levels.** Each parameter now carries a logging-sensitivity level — `none` (log plainly), `secure` (plain or masked depending on a runtime flag in your code), `strict` (always masked), `omit` (never logged) — exposed to templates as `Property.SecurityLevel`.
  - **How it works:** you assign levels in config — globally by name (`ParameterSecurityMappings`), as a project-wide default (`DefaultParameterSecurityLevel`, defaults to `secure`), or per function (`Functions[].Parameters[].SecurityLevel`, highest precedence). db-gen resolves the effective level and hands it to your template; **your template renders the masking**, so the tool stays language-agnostic (same philosophy as copy targets). A template branches on `$p.SecurityLevel` to emit `"******"`, a runtime-flag ternary, or the raw value — and drops `omit` params from the log line entirely. See [docs/templating.md → parameter security levels](./docs/templating.md#parameter-security-levels).
- **Example templates.** New top-level [`examples/`](./examples) with anonymized, real-world-shaped starting points: a C# `DbContext`/`Model`/`Processor` set, a `common-provider` that turns `SecurityLevel` into safe log lines, and a per-routine TypeScript model. See [examples/README.md](./examples/README.md).
- **Metadata contract test.** A golden-file test (`test/e2e/templates/metadata.gotmpl` → `test/e2e/golden/metadata/contract.txt`) dumps *every* field db-gen exposes to templates, so any change to the template-facing data model shows up as a reviewable diff.
- **`validate` command.** `db-gen validate` checks your configuration and templates without generating anything: it verifies settings are consistent, parses every template, and — when a database is reachable — renders each against your real routines to catch field/method typos. Exits non-zero on error, so it drops straight into CI. Use `--offline` to skip the database and render step. See [docs/usage.md → validate](./docs/usage.md#validate).
- **`--llm` reference.** `db-gen --llm` (or `db-gen llm`) prints one self-contained plain-text reference — concepts, every command and flag, the full config-key reference, the template data model, and template functions — for pasting into an LLM/AI coding assistant so it gets the whole tool in one shot. No database or config needed. See [docs/usage.md → llm](./docs/usage.md#llm).
- **Bare invocation prints version.** Running `db-gen` with no arguments now prints build/version information above the help output.

### v0.7.0

- **Copy targets — generic bulk-`COPY` code generation.** Declare a staging table and db-gen emits the metadata (ordered columns, type, nullability, context/data split, format hint) a per-language template needs to write a correct `COPY ... FROM STDIN`. No language specifics live in the tool. Schema drift on those tables is now reported in `generate` and `database-changes` output, so a changed staging column is no longer silent. See [docs/copy-targets.md](./docs/copy-targets.md).
- **Test framework.** A three-layer suite (pure unit, DB-backed integration, golden-file e2e) plus `make test` / `test-unit` / `test-integration` / `test-force` / `test-update-golden`. See [test/README.md](./test/README.md).
- **Security & toolchain (v0.7.1–v0.7.2).** Upgraded `pgx`/`x/text` and the Go toolchain to 1.25 to clear all `govulncheck` findings.

See [CHANGELOG.md](./CHANGELOG.md) for the full history.

## Docs

- 📘 [Usage & CLI reference](./docs/usage.md) — commands, flags, config resolution, offline workflow, change detection.
- ⚙️ [Configuration reference](./docs/configuration.md) — every config key, defaults, and what it does.
- 🧩 [Templating](./docs/templating.md) — template data model, functions, per-routine overrides, type mapping.
- 📦 [Copy targets](./docs/copy-targets.md) — bulk-`COPY` metadata contract and example templates.
- 🔐 [Context mapping & additional generators](./docs/context-mapping.md)
- ✅ [Validation](./docs/validation.md)
- 📚 [Examples / cookbook](./docs/examples.md) — config recipes for common setups.
- 🎁 [Example templates](./examples) — anonymized, real-world-shaped C# & TypeScript templates to copy from.
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
| `validate` | Check the config and templates (parse + render) without generating files. Exits non-zero on error. |
| `routines [out]` | Export routine metadata to JSON for offline generation. |
| `database-changes` | Show what changed in the database since the last generation. |
| `llm` | Print a full CLI reference (concepts, commands, config, template model) for LLM/AI assistants. Same as `--llm`. |
| `completion [shell]` | Emit a shell-completion script (`bash`, `zsh`, `fish`, `powershell`). |
| `version` | Print version / build information. |

Global flags: `--config`/`-s` (config file path), `--connectionString`/`-c`, `--debug`/`-d`, `--llm` (print the LLM reference). See [docs/usage.md](./docs/usage.md).

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
