# Templating

db-gen output is generated entirely from [Go templates](https://pkg.go.dev/text/template) you control. The tool stays language-agnostic — the template decides the syntax. This page documents the data each template receives, the helper functions available, and how type mapping and per-routine overrides work.

## Core templates

Configured in `db-gen.json`:

- **`DbContextTemplate`** — generates the database-call code (receives all routines at once).
- **`ModelTemplate`** — generates a model per routine return shape.
- **`ProcessorTemplate`** — generates a mapper from the db reader to the model, per routine.

For anything else (TypeScript types, providers, …) use [Additional Generators](./context-mapping.md). For bulk-`COPY` code, see [copy targets](./copy-targets.md).

## Template data

> Source of truth: `private/dbGen/types.go`.

```go
// DbContext template receives all routines:
type DbContextData struct {
    Config    *Config
    Functions []Routine
    BuildInfo *version.BuildInformation
}

// Model and Processor templates receive one routine each:
type ModelTemplateData struct {
    Config    *Config
    Routine   Routine
    BuildInfo *version.BuildInformation
}
type ProcessorTemplateData struct {
    Config    *Config
    Routine   Routine
    BuildInfo *version.BuildInformation
}

type Routine struct {
    FunctionName       string
    DbFullFunctionName string
    ModelName          string
    ProcessorName      string
    Schema             string
    DbFunctionName     string
    HasReturn          bool
    IsProcedure        bool
    Parameters         []Property   // all parameters, in order
    ReturnProperties   []Property
    UsesUserContext    bool
    ContextParameters  []Property   // params mapped to the context object
    RegularParameters  []Property   // params not mapped to context
}

type Property struct {
    DbColumnName       string
    DbColumnType       string
    PropertyName       string
    PropertyType       string // resolved type (accounts for nullable/optional)
    BaseType           string // base non-nullable type
    NullableReturnType string // type for nullable return values / model props
    NullableParamType  string // type for nullable parameters
    OptionalParamType  string // type for optional parameters (with DEFAULT)
    Position           int
    MapperFunction     string
    Nullable           bool   // can be unreliable
    Optional           bool   // parameters only: has a DEFAULT value
    IsContextParameter bool
    ContextPath        string
    ValidationRules    []ValidationRule
}
```

`Config` is the full [configuration object](./configuration.md), and `BuildInfo` carries version/build metadata — both available in every template.

## Template functions

Case conversion (default output is camelCase):

| Function | Result |
|----------|--------|
| `pascalCased` | `GetUserById` |
| `camelCased` | `getUserById` |
| `snakeCased` | `get_user_by_id` |
| `normalizeStr` | normalized db name with leading underscores stripped (e.g. `_user_id` → `user_id`); preserves underscores before numbers like `country_iso_2` |
| `trimPrefix` | strips a given prefix from a string |

Example:

```gotemplate
{{pascalCased $func.FunctionName}}
```

> Since v0.6.1, the bundled Model and Processor templates use `{{normalizeStr $property.DbColumnName}}` rather than `{{snakeCased $property.PropertyName}}`, to avoid information loss (e.g. `country_iso_2` collapsing to `country_iso2`) from round-tripping through PascalCase.

## Three-tier type mapping

A PostgreSQL type can resolve to up to four target types depending on context. Define them in [`Mappings`](./configuration.md#mappings--type-mapping):

- `MappedType` — base, non-nullable, non-optional.
- `NullableReturnType` — nullable return value / model property (e.g. `int?`).
- `NullableParameterType` — nullable parameter without a DEFAULT (e.g. `int?`).
- `OptionalParameterType` — parameter with a DEFAULT (e.g. `Optional<int>`).

Resolution (in `mapper.go`):

- **Parameters**: check `isOptional` first → use `OptionalParameterType` if set; else if nullable → `NullableParameterType` if set; otherwise `MappedType`.
- **Return/model columns**: nullable → `NullableReturnType` if set; otherwise `MappedType`.

**Nullable vs optional** is a real distinction:

- *Nullable* (no DEFAULT, accepts NULL) → always passed to the db, may be null → `T?`.
- *Optional* (has DEFAULT) → may be omitted from the call entirely → `Optional<T>`.

Templates typically check `$parameter.Optional` to decide between an "omit if absent" path (`.ToObjectOptional()` / `Optional.Some(...)`) and always passing the value.

Use `"*"` as a `DatabaseTypes` entry to provide a catch-all fallback for unmapped types; without it, an unmapped type stops generation with an error.

## Per-routine overrides

In `Generate[].Functions`, a function key can map to either a boolean (generate or not) or an override object. A key may be just the name (`my_func`) or include a signature (`my_func(text,int)`).

Override object fields:

- **`MappedName`** — overrides the generated name; model/processor names are derived by appending `Model`/`Processor`.
- **`DontRetrieveValues`** — disable value selection for a function that has return values (can only turn retrieval off, not on).
- **`SelectOnlySpecified`** — in `Model`, select only the columns you explicitly list.
- **`Model`** — object keyed by db column name. `false` skips the column; `true` or an object selects it. The object may override `MappedName`, `IsNullable`, `MappedType`, and `MappingFunction`.
  - Setting only `MappedType` looks up the mapping function from global mappings (errors if none found).
  - Setting only `MappingFunction` without `MappedType` does nothing.
- **`Parameters`** — object keyed by parameter name. Because you can't meaningfully select a subset of parameters, values must be objects (not booleans). You may override `MappedName`, `MappedType`, `IsNullable`, and `IsOptional`.

> When overriding types per function, leave fields you don't want to change as empty strings so the global mapping is looked up correctly — don't hardcode placeholder text.

## Overloaded functions

To avoid ambiguity, every function that is overloaded **must** be given a unique `MappedName`. The name must be unique within the schema (uniqueness is not yet enforced automatically, so be careful).
