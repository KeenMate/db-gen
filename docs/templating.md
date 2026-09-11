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
    SecurityLevel      string // logging sensitivity: none | secure | strict | omit
}
```

`Config` is the full [configuration object](./configuration.md), and `BuildInfo` carries version/build metadata — both available in every template.

> **Runnable examples:** minimal, test-covered templates live in [`test/templates/`](../test/templates) and [`test/e2e/templates/`](../test/e2e/templates). For fuller, real-world-shaped (anonymized) starting points — a C# DbContext/Model/Processor set, a secure-logging provider, and a TypeScript model — see [`examples/`](../examples).

## Template functions

Case conversion (default output is camelCase):

| Function | Result |
|----------|--------|
| `pascalCased` | `GetUserById` |
| `camelCased` | `getUserById` |
| `snakeCased` | `get_user_by_id` |
| `normalizeStr` | normalized db name with leading underscores stripped (e.g. `_user_id` → `user_id`); preserves underscores before numbers like `country_iso_2` |
| `trimPrefix` | strips a given prefix from a string |
| `templateVar` | looks up a [`TemplateVariables`](#template-variables) value by key (case-insensitive) |

Example:

```gotemplate
{{pascalCased $func.FunctionName}}
```

> Since v0.6.1, the bundled Model and Processor templates use `{{normalizeStr $property.DbColumnName}}` rather than `{{snakeCased $property.PropertyName}}`, to avoid information loss (e.g. `country_iso_2` collapsing to `country_iso2`) from round-tripping through PascalCase.

## Template variables

`TemplateVariables` is a free-form string→string map in the [config](./configuration.md#template-variables). It lets one shared template set serve multiple projects that differ only in config — e.g. output namespaces. Read a value with the `templateVar` function:

```gotemplate
namespace {{templateVar .Config.TemplateVariables "GeneratedNs"}};
```

Lookup is **case-insensitive**. Config loading lowercases every key, so `templateVar` lowercases the requested key before matching — author keys in whatever casing reads best (`GeneratedNs`, `generated_ns`, …) and reference them the same way. Values keep their exact case. A missing key yields an empty string.

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
- **`Parameters`** — object keyed by parameter name. Because you can't meaningfully select a subset of parameters, values must be objects (not booleans). You may override `MappedName`, `MappedType`, `IsNullable`, `IsOptional`, and `SecurityLevel`.

> When overriding types per function, leave fields you don't want to change as empty strings so the global mapping is looked up correctly — don't hardcode placeholder text.

## Parameter security levels

Every parameter carries a `SecurityLevel` describing how sensitive it is to logging. db-gen only resolves and exposes the level — it never renders masking itself. Your template decides what to emit for each level (this keeps db-gen language-agnostic, like [copy targets](./copy-targets.md)).

The four levels:

| Level | Meaning |
|-------|---------|
| `none` | Safe to log as plain text. |
| `secure` | Log as plain text or masked (`******`) depending on a **runtime** flag in your generated code (the issue calls it `db-gen:insecure-logging`). db-gen does not read this flag — your template emits code that does. |
| `strict` | Always log masked. |
| `omit` | Never include in any log output. |

Resolution precedence (in `mapper.go`): per-function `Parameters[name].SecurityLevel` → global [`ParameterSecurityMappings`](./configuration.md#parameter-security) by name → `DefaultParameterSecurityLevel` (defaults to `secure`). Matching is case-insensitive on the db parameter name.

Example template fragment (C#-flavored), branching on the level:

```gotemplate
{{range $p := .Routine.RegularParameters}}
{{- if ne $p.SecurityLevel "omit"}}
    log.Add("{{$p.PropertyName}}",
    {{- if eq $p.SecurityLevel "strict"}} "******"
    {{- else if eq $p.SecurityLevel "secure"}} InsecureLogging ? {{$p.PropertyName}}.ToString() : "******"
    {{- else}} {{$p.PropertyName}}.ToString()
    {{- end}});
{{- end}}
{{end}}
```

## Overloaded functions

When two or more functions share a name within a schema (PostgreSQL overloads), db-gen keeps their generated names unique automatically:

- Each member of the overload set gets a **stable, 1-based numeric suffix** appended to its generated name — e.g. `get_something(int4)` → `GetSomething1`, `get_something(text)` → `GetSomething2` (with matching `GetSomething1Model` / `GetSomething2Processor` and file names).
- The suffix order is **deterministic**: overloads are sorted by their full parameter signature, so the same database always produces the same suffixes regardless of physical row order or the machine running db-gen. Adding a new overload whose signature sorts earlier can shift the numbers, so pin names with `MappedName` if you need them frozen. db-gen logs a warning for any overload set that has no `MappedName` at all, so the shift risk is visible in CI output.
- A per-overload **`MappedName` overrides the suffix** — set one (keyed by the full signature, e.g. `"get_something(text)": { "MappedName": "GetSomethingByName" }`) and that exact name is used with no suffix. `MappedName` values are not checked for uniqueness, so make sure they don't collide.
