# Feature inventory — db-gen

A complete, categorized list of what db-gen exposes. Each row names the **exact** identifier from the source (CLI command/flag, config key, or template symbol), its default, and whether it has automated test coverage.

How to read the matrix:

- **Surface** — how the feature is used: `cmd` (CLI command), `flag` (CLI flag), `config` (config key), `config[]` (array-of-objects field), `template-fn` (template function), `template-data` (field on template data).
- **Identifier** — the literal name. Config keys are PascalCase as in `db-gen.json`; flags are `--long`/`-short`.
- **Default** — for booleans, all default `false` unless stated.
- **Tested** — the test file(s) covering it, or `manual` / `n/a`. See [test/README.md](./test/README.md).

> Source of truth: `Config` in `private/dbGen/config.go`, template data in `private/dbGen/types.go`, commands in `cmd/`. Re-check after each release.

---

## 1. Commands

| Feature | Surface | Identifier | Notes | Tested |
|---|---|---|---|---|
| Generate code | cmd | `generate` | Connect/load, map, render templates; reports changes | `e2e_test.go` |
| Validate config & templates | cmd | `validate` | Static settings checks + template parse + render (if DB reachable); `--offline` skips DB. Exits non-zero on error | `validate_test.go` |
| Export routines (offline) | cmd | `routines [out]` | Writes routine metadata JSON | `integration_test.go` |
| Show schema changes | cmd | `database-changes` | Diff only, no generation | `changedetection_test.go` |
| LLM reference | cmd | `llm` | Prints full CLI reference (concepts/commands/config/template model); same as `--llm`. No DB/config needed | n/a |
| Shell completion | cmd | `completion [shell]` | `bash`/`zsh`/`fish`/`powershell` | n/a |
| Version info | cmd | `version` | Prints build info | n/a |

## 2. Global flags

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Config file path | flag | `--config` / `-s` | — | Else default locations tried | `config_test.go` |
| Connection string | flag | `--connectionString` / `-c` | — | Overrides config value | manual |
| Debug logging | flag | `--debug` / `-d` | off | Debug logs + debug files | manual |
| LLM reference | flag | `--llm` | off | Prints the full CLI reference, then exits; same as the `llm` command | n/a |

## 3. Connection & output

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Connection string | config | `ConnectionString` | — | PostgreSQL DSN | `integration_test.go` |
| Output folder | config | `OutputFolder` | — | Relative to config dir | `config_test.go` |
| Models subfolder | config | `ModelsFolderName` | `"models"` | | `e2e_test.go` |
| Processors subfolder | config | `ProcessorsFolderName` | `"processors"` | | n/a |
| File extension | config | `GeneratedFileExtension` | — | e.g. `.cs` | `e2e_test.go` |
| File case | config | `GeneratedFileCase` | — | `snakecase`/`camelcase`/`pascalcase`; validated | `config_test.go`, `generator_test.go` |
| Clear output folder | config | `ClearOutputFolder` | `false` | Empties output before gen | `generator_test.go` |
| Remove orphaned files | config | `RemoveOrphanedFiles` | `false` | Deletes files for vanished functions; needs `ClearOutputFolder:false` | `generator_test.go` |

## 4. Generation outputs

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| DbContext template | config | `DbContextTemplate` | — | Database-call code | `e2e_test.go` |
| Generate models | config | `GenerateModels` | `false` | | `e2e_test.go` |
| Model template | config | `ModelTemplate` | — | | `e2e_test.go` |
| Template variables | config | `TemplateVariables` | — | String map read via `templateVar`; keys lowercased on load | `config_test.go` |
| Generate processors | config | `GenerateProcessors` | `false` | | n/a |
| Processor template | config | `ProcessorTemplate` | — | | n/a |
| Processors for void returns | config | `GenerateProcessorsForVoidReturns` | `false` | | n/a |

## 5. Type mapping

| Feature | Surface | Identifier | Notes | Tested |
|---|---|---|---|---|
| Type mappings | config[] | `Mappings` | List; last wins on conflict; `"*"` = fallback | `mapper_test.go`, `preprocess_test.go` |
| Database types covered | config | `Mappings[].DatabaseTypes` | `udt_name`s; `"*"` catch-all | `mapper_test.go` |
| Base type | config | `Mappings[].MappedType` | Non-null, non-optional | `mapper_test.go` |
| Reader function | config | `Mappings[].MappingFunction` | | `mapper_test.go` |
| Nullable return type | config | `Mappings[].NullableReturnType` | e.g. `int?` | `mapper_test.go` |
| Nullable parameter type | config | `Mappings[].NullableParameterType` | param, no DEFAULT | `mapper_test.go` |
| Optional parameter type | config | `Mappings[].OptionalParameterType` | param, has DEFAULT | `mapper_test.go` |

## 6. Function & schema selection

| Feature | Surface | Identifier | Notes | Tested |
|---|---|---|---|---|
| Schema selection | config[] | `Generate[].Schema` | | `filter_test.go` |
| All functions | config | `Generate[].AllFunctions` | Except those set `false` | `filter_test.go` |
| Per-function selection | config | `Generate[].Functions` | name or `name(args)` → bool/object | `filter_test.go` |
| Rename routine | config | `Functions[].MappedName` | Required for overloads | `mapper_test.go` |
| Disable value retrieval | config | `Functions[].DontRetrieveValues` | Off-only | n/a |
| Select only specified columns | config | `Functions[].SelectOnlySpecified` | | n/a |
| Per-column model override | config | `Functions[].Model` | `MappedName`/`MappedType`/`MappingFunction`/`IsNullable` | n/a |
| Per-parameter override | config | `Functions[].Parameters` | `MappedName`/`MappedType`/`IsNullable`/`IsOptional` | `mapper_test.go` |
| Overloaded function handling | behavior | (mark overloaded) | Forces unique `MappedName` | `preprocess_test.go` |

## 7. Context parameters

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Enable context | config | `UseUserContext` | `false` | | n/a |
| Context param name | config | `UserContextParameterName` | `"ctx"` | | n/a |
| Context type | config | `UserContextType` | `"UserContext"` | | n/a |
| Context mappings | config[] | `ContextParameterMappings` | — | name(s) → context path | `mapper_test.go` |
| Mapped parameter names | config | `ContextParameterMappings[].ParameterNames` | — | | `mapper_test.go` |
| Context path | config | `ContextParameterMappings[].ContextPath` | — | | `mapper_test.go` |

## 7a. Parameter security (logging sensitivity)

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Default security level | config | `DefaultParameterSecurityLevel` | `"secure"` | fallback level; `none`/`secure`/`strict`/`omit` | `mapper_test.go` |
| Global security mappings | config[] | `ParameterSecurityMappings` | — | name(s) → level | `mapper_test.go` |
| Mapped parameter names | config | `ParameterSecurityMappings[].ParameterNames` | — | case-insensitive | `mapper_test.go` |
| Security level | config | `ParameterSecurityMappings[].SecurityLevel` | — | | `mapper_test.go` |
| Per-function override | config | `Generate[].Functions[].Parameters[].SecurityLevel` | — | wins over global | `mapper_test.go` |
| Resolved level (template) | template-data | `Property.SecurityLevel` | — | `none`/`secure`/`strict`/`omit` | `mapper_test.go`, `e2e_test.go` |

## 8. Additional generators

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Additional generators | config[] | `AdditionalGenerators` | — | Custom outputs | manual |
| Enable / name | config | `…[].Enabled` / `.Name` | | | manual |
| Template / output | config | `…[].Template` / `.OutputFolder` | | | manual |
| Single-file file name | config | `…[].FileName` | | Implies `single-file` | manual |
| Per-routine extension | config | `…[].FileExtension` | | Implies `per-routine` | manual |
| File case | config | `…[].FileCase` | | | manual |
| Generation type | config | `…[].GenerationType` | inferred | `single-file`/`per-routine` | manual |
| Clean output folder | config | `…[].CleanOutputFolder` | `false` | | manual |

## 9. Copy targets

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Enable copy targets | config | `GenerateCopyTargets` | `false` | | `copytargets_test.go` |
| Copy template | config | `CopyTargetTemplate` | — | | `e2e_test.go` |
| Copy subfolder | config | `CopyTargetsFolderName` | `"copy"` | | `copytargets_test.go` |
| Target tables | config[] | `CopyTargets` | — | | `copytargets_test.go` |
| Schema / table | config | `CopyTargets[].Schema` / `.Table` | — | | `integration_test.go` |
| Mapped name | config | `CopyTargets[].MappedName` | — | | `copytargets_test.go` |
| Format hint | config | `CopyTargets[].Format` | `"csv"` | `csv`/`text`/`binary` | `copytargets_test.go` |
| NULL sentinel | config | `CopyTargets[].NullString` | — | text/CSV | n/a |
| Context/data column split | template-data | `CopyTarget.ContextColumns`/`.DataColumns`/`.AllColumns` | — | wire order: context then data | `copytargets_test.go` |

## 10. Offline & change detection

| Feature | Surface | Identifier | Default | Notes | Tested |
|---|---|---|---|---|---|
| Routines file path | config | `RoutinesFile` | `"./db-gen-routines.json"` | | n/a |
| Use routines file | config | `UseRoutinesFile` | `false` | Offline generation | n/a |
| Routine change detection | behavior | (generation info diff) | — | created/deleted/renamed/type/nullable params | `changedetection_test.go` |
| Copy-target change detection | behavior | (table diff) | — | created/deleted tables, column add/remove/type/nullable | `changedetection_test.go` |

## 11. Templating

| Feature | Surface | Identifier | Notes | Tested |
|---|---|---|---|---|
| DbContext data | template-data | `DbContextData{Config,Functions,BuildInfo}` | All routines | `e2e_test.go` |
| Model/Processor data | template-data | `ModelTemplateData`/`ProcessorTemplateData` | One routine | `e2e_test.go` |
| Routine fields | template-data | `Routine{Parameters,ReturnProperties,ContextParameters,RegularParameters,…}` | | `mapper_test.go` |
| Property fields | template-data | `Property{BaseType,Nullable,Optional,Position,SecurityLevel,…}` | | `mapper_test.go` |
| Pascal case | template-fn | `pascalCased` | | `strings_test.go` |
| Camel case | template-fn | `camelCased` | | `strings_test.go` |
| Snake case | template-fn | `snakeCased` | | `strings_test.go` |
| Normalize string | template-fn | `normalizeStr` | Strips leading underscores; keeps `_` before digits | `strings_test.go` |
| Trim prefix | template-fn | `trimPrefix` | | n/a |
| Template variable lookup | template-fn | `templateVar` | Case-insensitive lookup into `Config.TemplateVariables` | `generator_test.go` |

## 12. Validation

| Feature | Surface | Identifier | Notes | Tested |
|---|---|---|---|---|
| Validation config | config | `Validation` | Reusable parameter rules — see [docs/validation.md](./docs/validation.md) | manual |
| Strategy | config | `Validation.ValidationStrategy` | FluentValidation/Manual/… | manual |
| Rule definitions | config[] | `Validation.ValidationRuleDefinitions` | | manual |
| Parameter→rule mappings | config[] | `Validation.ParameterValidationMappings` | | manual |
| Function-specific overrides | config | `Validation.FunctionSpecificValidations` | | manual |

---

### Cross-references

- **CLI & config resolution:** [docs/usage.md](./docs/usage.md)
- **Full config reference:** [docs/configuration.md](./docs/configuration.md)
- **Template data & functions:** [docs/templating.md](./docs/templating.md)
- **Copy targets:** [docs/copy-targets.md](./docs/copy-targets.md)
- **Context mapping & generators:** [docs/context-mapping.md](./docs/context-mapping.md)
- **Validation:** [docs/validation.md](./docs/validation.md)
- **Test framework:** [test/README.md](./test/README.md)
