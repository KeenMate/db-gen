# CHANGELOG

## 0.7.1

### Security
- Upgraded dependencies to clear all reachable vulnerabilities reported by `govulncheck`:
  - `github.com/jackc/pgx/v5` v5.5.1 → v5.10.0 (fixes `GO-2024-2606` SQL injection, `GO-2024-2567` Pipeline panic, `GO-2026-5004`).
  - `golang.org/x/text` v0.14.0 → v0.40.0.
  - Bumped Go toolchain to 1.25 in `go.mod` and CI, clearing 16 standard-library advisories (TLS/x509/url/pem) in released binaries.

## 0.7.0

### New Features
- **Copy targets** — generate language-agnostic bulk-`COPY`-into-table code from a table's columns. db-gen emits column metadata (ordered columns, type, nullability, context/data split, format hint) and a per-language template produces the `COPY ... FROM STDIN` code. See [docs/copy-targets.md](./docs/copy-targets.md).
  - New config: `GenerateCopyTargets`, `CopyTargetTemplate`, `CopyTargetsFolderName`, `CopyTargets[]` (`Schema`, `Table`, `MappedName`, `Format`, `NullString`).
  - New template data: `CopyTargetTemplateData` / `CopyTarget` with `ContextColumns` / `DataColumns` / `AllColumns`.
  - Example templates added: `copy-pgx.gotmpl` (Go/pgx) and `copy-npgsql.gotmpl` (C#/Npgsql).
- **Copy-target change detection** — staging-table schema changes (added/removed columns, type and nullability changes) are now reported by `generate` and `database-changes`, closing the gap where staging-column drift was silent.
- **Test framework** — three-layer suite (pure unit, DB-backed integration, golden-file e2e) with `make test` / `test-unit` / `test-integration` / `test-force` / `test-update-golden`. See [test/README.md](./test/README.md).

### Documentation
- README slimmed into a landing page; reference docs split into `docs/` (`usage`, `configuration`, `templating`, `copy-targets`, `context-mapping`, `validation`, `examples`).
- Added `FEATURES.md` feature-inventory matrix.

## 0.6.1

### Bug Fixes
- **Fixed field name generation**: Model and processor field names now preserve underscores before numbers (e.g., `country_iso_2` instead of `country_iso2`)
  - Added `normalizeStr` template function to directly use normalized database column names
  - Prevents information loss from round-trip conversion through PascalCase
  - Maintains consistency with original ecto_gen behavior
- **Fixed unnamed parameter handling**: Functions with unnamed parameters (e.g., `$1`, `$2`) no longer cause errors during routine loading
  - `parameter_name` is now coalesced to empty string when NULL

### Template Changes
- Model template now uses `{{normalizeStr $property.DbColumnName}}` instead of `{{snakeCased $property.PropertyName}}`
- Processor template now uses `{{normalizeStr $property.DbColumnName}}` instead of `{{snakeCased $property.PropertyName}}`
- New template function `normalizeStr` available for removing leading underscores from strings

## 0.6.0

### New Features
- **Context Parameter Mapping**: Automatically inject user context parameters (user_id, created_by, tenant_id, etc.) from UserContext object
- **Additional Generators Framework**: Extensible system to generate additional outputs (CommonProvider, TypeScript models, etc.) using custom templates
- **Three-tier Type Mapping System**: Separate type mappings for base types, nullable return types, nullable parameters, and optional parameters
- **CleanOutputFolder Option**: Clean output folder before generation for AdditionalGenerators
- **RemoveOrphanedFiles Option**: Automatically remove generated files when their corresponding database functions are deleted

### Bug Fixes
- Fixed per-function type override to properly use global nullable/optional type mappings
- Fixed nullable vs optional parameter handling - nullable parameters now use `T?` instead of `Optional<T>` to ensure they're always passed to database
- Fixed double-wrapping issue where nullable parameters were being wrapped in `Optional.Some()` when already typed as `Optional<T>`
- Fixed jsonb parameter handling to prevent PostgreSQL-specific types from leaking outside DbContext layer
- Fixed change detection to track files from all output folders (main + AdditionalGenerators), not just main output folder

### Template Changes
- DbContext template now checks only `$parameter.Optional` (not `$parameter.Nullable`) for `.ToObjectOptional()` usage
- Added jsonb-to-JsonbStringParameter conversion logic in DbContext template for architectural boundary maintenance
- Templates now have access to `BaseType`, `NullableReturnType`, `NullableParamType`, and `OptionalParamType` on Property objects

### Configuration Changes
- Added `UseUserContext`, `UserContextParameterName`, `UserContextType` options
- Added `ContextParameterMappings` for automatic context parameter injection
- Added `AdditionalGenerators` with support for `single-file` and `per-routine` generation types
- Type mappings now support `NullableReturnType`, `NullableParameterType`, and `OptionalParameterType` fields
- Added `CleanOutputFolder` boolean option to AdditionalGenerator configuration
- Added `RemoveOrphanedFiles` boolean option to automatically clean up files for deleted database functions

## 0.5.2

### New features
- Separate Optional from Nullable
- Optional is only used in parameters and specifies that given parameter has default value
- Both can be override separately

## 0.5.0

### Breaking changes:
- Changed configuration format for functions, instead of array accept object where keys are function names
- Remove `IgnoredFunctions` in favor of setting function to false in `Functions`
- Enforce mapping for overloaded function

### New Features

#### Mapping per function
- You can now specify mapping per function that will override global settings
- Change mapped function name
- Disable fetching of values
- Change mapped type and is nullable for every parameter and return model column separately
- Set custom mapping function when changing type, looks at global mapping if you do not define any

## 4.0.0

### New features

- Add new command `routines` loads routines from database and saves them to file set in `RoutinesFile`
- added `--UseRoutinesFile` flag to generate, which loads routes from `RoutinesFile` instead of database

## 0.3.4

### New features

- local config can have both `.local` postfix

## 0.3.1

### New features

- local config can have both `.local.` or `local.` as prefix

## 0.3.1

### New features

- Build information is now available in all templates `.BuildInfo`

## 0.3.0

### Breaking changes!

- `GeneratedFileCase` config values renamed to match world-wide accepted terms:
	- from: `"snake"` to: `"snakecase"`
	- from: `"lcase"` to: `"camelcase"`
	- from: `"ucase"` to: `"pascalcase"`

- Template functions: `uCamel`, `lCamel`, `snake` renamed with equivalent world-wide accepted terms:
	- from: `"snake"` to: `"snakeCased"`
	- from: `"lCamel"` to: `"camelCased"`
	- from: `"uCamel"` to: `"pascalCased"`

- Model template now has previous config values accessible through `Routine` variable

### New features

- New config values for folder names where generated models and processors will be placed
	- `ProcessorsFolderName` with default value: `"processors"`
	- `ModelsFolderName` with default value: `"models"`
- Model, Processor and DbContext templates now have new variable `Config` pointing to the config values of the application
