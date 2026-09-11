package cmd

import (
	"fmt"

	"github.com/keenmate/db-gen/private/version"
	"github.com/spf13/cobra"
)

// llmCmd prints a single, self-contained reference document describing db-gen's
// concepts, commands, flags, and configuration. It is aimed at LLM/AI coding
// assistants that need the full mental model of the tool in one shot, without
// spelunking source or docs. The same output is available as the `--llm` flag.
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Print a full CLI reference for LLM/AI assistants",
	Long: `Print one self-contained reference document (concepts, commands, flags,
and configuration keys) intended for LLM/AI coding assistants.

Equivalent to running db-gen --llm.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(formatLlmOutput())
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
}

// formatLlmOutput returns the full plain-text reference document.
func formatLlmOutput() string {
	return fmt.Sprintf(`db-gen %s — code generator for PostgreSQL stored functions and procedures
by Keenmate s.r.o. | https://github.com/keenmate/db-gen

db-gen connects to a PostgreSQL database, reads metadata for stored functions
and procedures ("routines"), and generates typed database-access code from Go
templates that you control. It is language-agnostic and NOT an ORM: you write
SQL functions, db-gen writes the typed call/mapping code around them. Emphasis
is on consistency, customization, and offline operation.

Use "db-gen --llm" or "db-gen llm" to print this document.

CONCEPTS

Routines:
  A routine is a stored function or procedure read from PostgreSQL. db-gen reads
  its metadata, applies schema/function filters and type mappings, then renders
  templates. Each routine has:
    Parameters       — inputs; may be mandatory, nullable, or have a DEFAULT.
    ReturnProperties — columns returned (for table-returning functions).

Nullable vs Optional (a core distinction):
  Nullable — parameter has no DEFAULT but accepts NULL. Must always be passed to
    the DB; can be null. Maps to NullableParameterType (e.g. int?).
  Optional — parameter has a DEFAULT, so it can be omitted from the DB call.
    Maps to OptionalParameterType (e.g. Optional<int>).
  Templates check $parameter.Optional to decide how to pass the value.

Overloaded functions:
  When several functions share a name in one schema (PostgreSQL overloads),
  db-gen keeps their generated names unique by appending a stable, 1-based
  numeric suffix (get_something(int4) -> GetSomething1, get_something(text) ->
  GetSomething2, plus matching Model/Processor/file names). Overloads are sorted
  by full parameter signature, so suffixes are deterministic across runs and
  machines. A per-overload MappedName (keyed by full signature, e.g.
  "get_something(text)") overrides the suffix. db-gen warns when an overload set
  has no MappedName at all, since positional suffixes can shift if signatures
  change.

Templates (three core types):
  DbContext  — receives ALL processed routines at once; generates the main
               database-call facade. Data: { Config, Functions []Routine, BuildInfo }.
  Model      — receives ONE routine; generates the return-shape model/DTO.
               Data: { Config, Routine, BuildInfo }.
  Processor  — receives ONE routine; generates the mapper from a DB reader to a
               model. Data: { Config, Routine, BuildInfo }.
  Templates are Go text/template (.gotmpl) files. Output language/framework is
  entirely determined by the templates you supply.

Additional Generators:
  Extend generation beyond the three core templates (e.g. TypeScript types, a
  CommonProvider, WebSocket endpoints). Each generator is either "single-file"
  (one file for all routines) or "per-routine" (one file per function) and
  receives the same template data as the core templates.

Copy Targets:
  Generate bulk "COPY ... FROM STDIN" code for staging tables. db-gen stays
  language-agnostic and emits CopyTarget metadata; templates produce the COPY
  code. Requires a live database connection (not part of the offline routines
  file). Each target exposes:
    ContextColumns — columns matched against ContextParameterMappings by db
                     column name (case-insensitive); injected once per call.
    DataColumns    — real data columns in table-ordinal order (each has Position).
    AllColumns     — ContextColumns ++ DataColumns, in wire order (emit verbatim
                     as the COPY (col, col, ...) list).

Three-tier type mapping:
  Each PostgreSQL type can resolve to up to four target-language types by context:
    MappedType            — base type (non-nullable, non-optional), e.g. int
    NullableReturnType    — nullable return value / model property, e.g. int?
    NullableParameterType — nullable parameter with NO DEFAULT, e.g. int?
    OptionalParameterType — parameter WITH a DEFAULT, e.g. Optional<int>
  Parameter resolution (in mapper.go): check isOptional first (use
  OptionalParameterType if set), else if nullable use NullableParameterType,
  else use MappedType. Return resolution: nullable uses NullableReturnType if
  set, else MappedType. Per-function overrides can adjust MappedType,
  IsNullable, IsOptional.

Context parameter mapping:
  Set UseUserContext=true to inject parameters from a context object (e.g. user
  id, tenant id) instead of passing them at every call site. Map db parameter
  names to a context path (e.g. "User.UserId"). Templates then see
  $routine.ContextParameters and $routine.RegularParameters, and per-property
  IsContextParameter / ContextPath. Matching is by db parameter name,
  case-insensitive. Handled by processContextParameters() in mapper.go;
  copy-target equivalent matches by db column name.

Parameter security levels (secure logging):
  Each parameter carries a logging-sensitivity level; templates decide what to
  emit. Levels: none (log plain), secure (plain or masked at runtime), strict
  (always masked), omit (never logged). Resolution: per-function
  Parameters[param].SecurityLevel -> global ParameterSecurityMappings (by name,
  case-insensitive) -> DefaultParameterSecurityLevel (default "secure").

Offline / routines-file workflow:
  1. Online, once: "db-gen routines [out]" exports routine metadata to JSON
     (default ./db-gen-routines.json).
  2. Offline: set UseRoutinesFile=true (or pass --useRoutinesFile); "db-gen
     generate" reads the JSON instead of connecting to the database.

Change detection:
  db-gen records the routines and copy-target tables it generated against in a
  generation-information file in the output folder. On later runs, "generate"
  prints detected schema changes then regenerates; "database-changes" prints the
  same diff WITHOUT generating (useful for CI drift detection). Reported: created
  / deleted functions, renamed / added / removed parameters, parameter
  type/nullability changes, and copy-target table/column changes.

Config file resolution:
  Primary locations (first match wins):
    ./db-gen.json
    ./db-gen/db-gen.json
    ./db-gen/config.json
  Or pass an explicit path with --config/-s.
  Optional local override (deep-merged on top of the primary). For a primary
  named db-gen.json in the same directory, these are tried in order:
    local.db-gen.json
    .local.db-gen.json
    db-gen.local
    db-gen.json.local
  All relative paths inside the config (templates, output folder, routines file)
  resolve relative to the CONFIG FILE's directory, not the current directory.
  JSON and YAML config files are both supported (by file extension).

COMMANDS

db-gen generate
  Generate code. Loads routines from the database (or the routines file), detects
  changes, renders all enabled templates, and writes the generation-information
  file. Output folder and templates come from the config.
    --config, -s <path>            path to configuration file
    --connectionString, -c <str>   PostgreSQL connection string (overrides config)
    --debug, -d                    print debug logs and write debug files
    --useRoutinesFile              read routines from the routines file instead of the DB

db-gen routines [out]
  Export routine metadata from the database to a JSON file so code can be
  generated later offline. [out] overrides the output path (default: config
  RoutinesFile, else ./db-gen-routines.json). Always reads from the database.
    --config, -s <path>            path to configuration file
    --connectionString, -c <str>   PostgreSQL connection string (overrides config)
    --debug, -d                    print debug logs and write debug files

db-gen database-changes
  Print differences between the current database (or routines file) and the
  routines recorded at the last generation. Prints only; generates nothing.
  Exits with an error if no generation information is found.
    --config, -s <path>            path to configuration file
    --connectionString, -c <str>   PostgreSQL connection string (overrides config)
    --debug, -d                    print debug logs and write debug files
    --useRoutinesFile              read routines from the routines file instead of the DB

db-gen validate
  Validate the configuration and the templates it references without generating
  files. Checks, in order: (1) settings load and enabled features are configured
  consistently; (2) every used template parses; (3) if a DB is reachable (or
  UseRoutinesFile is set) and not --offline, each template is rendered against
  your real routines to catch field/method errors. Exits non-zero on any error.
    --config, -s <path>            path to configuration file
    --connectionString, -c <str>   PostgreSQL connection string (overrides config)
    --debug, -d                    print debug logs and write debug files
    --useRoutinesFile              use the routines file instead of a database connection
    --offline                      only validate settings and parse templates; never
                                   connect to a database or render templates

db-gen completion [bash|zsh|fish|powershell]
  Print a shell-completion script for the given shell. Sourcing instructions are
  included in the output.

db-gen version
  Print executable version and build information.

db-gen llm
  Print this reference document. Equivalent to "db-gen --llm".

db-gen help [command]
  Print help for a command.

Global flag:
  --llm    print this reference document, then exit.

CONFIGURATION KEYS

Booleans default to false unless noted. Defaults shown are db-gen's built-in
defaults. Relative paths resolve against the config file's directory.

Connection & output:
  ConnectionString        string   PostgreSQL connection string (best kept in a local override).
  OutputFolder            string   Folder for generated files. Created if missing.
  ModelsFolderName        string   Subfolder for models (default "models").
  ProcessorsFolderName    string   Subfolder for processors (default "processors").
  GeneratedFileExtension  string   Extension for generated files, e.g. ".cs", ".ts".
  GeneratedFileCase       string   Filename case: "snakecase" | "camelcase" | "pascalcase" (required).
  ClearOutputFolder       bool     Delete output-folder contents before generating.
  RemoveOrphanedFiles     bool     Remove previously generated files whose function no longer exists
                                   (only when ClearOutputFolder is false).
  Debug                   bool     Debug logging + debug files (usually set via --debug).

What to generate:
  GenerateModels                    bool   Generate models for routine return shapes.
  GenerateProcessors                bool   Generate processors (DB reader -> model mappers).
  GenerateProcessorsForVoidReturns  bool   Also generate a processor for void-returning functions.

Templates:
  DbContextTemplate   string   Template for the database-call facade.
  ModelTemplate       string   Template for model files.
  ProcessorTemplate   string   Template for processor files.

Template variables:
  TemplateVariables   object(string->string)   Free-form values, read in templates with
                      {{templateVar .Config.TemplateVariables "Key"}} (case-insensitive keys).

Selection & mapping:
  Generate   []SchemaConfig   Which schemas/functions to generate. Each:
               Schema        string
               AllFunctions  bool                       generate all functions in the schema...
               Functions     map[name]RoutineMapping    ...except those set false; or per-function overrides.
             RoutineMapping (per-function override): MappedName, DontRetrieveValues,
               SelectOnlySpecified, Model (map[dbColumn]ColumnMapping), Parameters (map[name]ParamMapping).
             ColumnMapping: MappedName, MappedType, MappingFunction, IsNullable.
             ParamMapping: MappedName, MappedType, IsNullable, IsOptional, SecurityLevel.
  Mappings   []Mapping        Type mappings. Each:
               DatabaseTypes []string   PostgreSQL udt_names (use "*" as catch-all fallback).
               MappedType string, MappingFunction string, NullableReturnType string,
               NullableParameterType string, OptionalParameterType string.
             Later entries win when a type appears more than once.

Offline generation:
  RoutinesFile      string   Path to the routines JSON (default "./db-gen-routines.json").
  UseRoutinesFile   bool     Read routines from RoutinesFile instead of the database.

Context parameters:
  UseUserContext             bool     Enable context-parameter injection.
  UserContextParameterName   string   Name of the context parameter in generated code (default "ctx").
  UserContextType            string   Type of the context parameter (default "UserContext").
  ContextParameterMappings   []{ ParameterNames []string; ContextPath string }
                             Map db parameter names (case-insensitive) to a context path, e.g. "User.UserId".

Parameter security:
  DefaultParameterSecurityLevel   string   Level for unmatched parameters (default "secure").
  ParameterSecurityMappings       []{ ParameterNames []string; SecurityLevel string }
                                  Global by-name levels: none | secure | strict | omit.

Additional generators:
  AdditionalGenerators   []AdditionalGenerator. Each:
    Name              string   display name for logging.
    Enabled           bool     whether to run this generator.
    Template          string   template path.
    OutputFolder      string   output folder.
    FileName          string   required for single-file generators.
    FileExtension     string   required for per-routine generators.
    FileCase          string   "snakecase" | "camelcase" | "pascalcase".
    GenerationType    string   "single-file" | "per-routine" (auto: single-file if FileName set, else per-routine).
    CleanOutputFolder bool     delete the output folder before generating.

Copy targets:
  GenerateCopyTargets     bool     Enable copy-target generation (needs a live DB).
  CopyTargetTemplate      string   Template for copy-target files.
  CopyTargetsFolderName   string   Subfolder for copy-target files (default "copy").
  CopyTargets             []CopyTargetConfig. Each:
    Schema string, Table string, MappedName string (override name),
    Format string ("csv" | "text" | "binary", template hint), NullString string (NULL sentinel).

Validation subsystem:
  Validation   ValidationConfig   Optional validation-code configuration:
    ValidationStrategy string, ValidationLocations []string,
    ValidationBehavior { CollectAllErrors bool (default true), ReturnType string
      (default "ValidationResult"), ErrorResponseTemplate string },
    ValidationRuleDefinitions [], ParameterValidationMappings [],
    FunctionSpecificValidations map[function]map[parameter][]rule.

TEMPLATE DATA (available in templates)

  Config      — the full Config object.
  Functions   — []Routine (DbContext template only).
  Routine     — one routine (Model/Processor templates): FunctionName, Schema,
                DbFullFunctionName, ModelName, ProcessorName, HasReturn,
                IsProcedure, Parameters, ReturnProperties, UsesUserContext,
                ContextParameters, RegularParameters.
  CopyTarget  — one copy target (copy template): Schema, Table, StructName,
                Format, NullString, ContextColumns, DataColumns, AllColumns.
  BuildInfo   — version and build metadata.
  Each Property (parameter / return column / copy column) exposes at least:
    DbColumnName, DbColumnType, PropertyName, PropertyType, Nullable, Optional,
    MapperFunction, Position, IsContextParameter, ContextPath, SecurityLevel.

TEMPLATE FUNCTIONS

  pascalCased s        -> PascalCase
  camelCased s         -> camelCase
  snakeCased s         -> snake_case
  normalizeStr s       -> normalized name (strips leading underscores)
  trimPrefix s prefix  -> strip a prefix
  templateVar map key  -> look up a TemplateVariables value (case-insensitive)

DOCS

  README.md, docs/usage.md, docs/configuration.md, docs/templating.md,
  docs/copy-targets.md, docs/context-mapping.md, docs/validation.md.`,
		versionLabel())
}

// versionLabel returns a printable version string for the reference header,
// falling back gracefully for local builds where no version is stamped.
func versionLabel() string {
	if version.IsLocalBuild() {
		return "(local build)"
	}
	return version.GetVersion()
}
