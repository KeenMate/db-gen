# CHANGELOG

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
