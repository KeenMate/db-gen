# CHANGELOG

# 0.2.2

## Breaking changes!

- `GeneratedFileCase` config values renamed to match world-wide accepted terms:
  - from: `"snake"` to: `"snakecase"`
  - from: `"lcase"` to: `"camelcase"`
  - from: `"ucase"` to: `"pascalcase"`

- Template functions: `uCamel`, `lCamel`, `snake` renamed with equivalent world-wide accepted terms:
	- from: `"snake"` to: `"snakeCased"`
	- from: `"lCamel"` to: `"camelCased"`
	- from: `"uCamel"` to: `"pascalCased"`

- Model template now has previous config values accessible through `Routine` variable

## New features

- New config values for folder names where generated models and processors will be placed
  - `ProcessorsFolderName` with default value: `"processors"`
  - `ModelsFolderName` with default value: `"models"`
- Model, Processor and DbContext templates now have new variable `Config` pointing to the config values of the application