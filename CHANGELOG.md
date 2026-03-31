# CHANGELOG

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
