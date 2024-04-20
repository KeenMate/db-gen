# CHANGELOG

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
