# Copy targets

Copy targets generate the boilerplate for bulk-loading rows into a staging table with PostgreSQL `COPY ... FROM STDIN` — the fast path for "stage everything, then process server-side in one call".

Like everything else in db-gen, this is **language-agnostic**: the tool emits *metadata* describing the table; a per-language template turns it into correct `COPY` code (text/CSV with escaping, or binary). No language specifics live in the tool.

It also closes a long-standing gap: changes to a staging table's columns are now reported by `generate` / `database-changes`, so a renamed or retyped staging column is no longer silent. See [change detection](./usage.md#change-detection).

## Why

Apps that bulk-import data tend to hand-write the same `COPY` plumbing per table: the column list in the right order, where the context columns (`created_by`, `job_run_id`, …) go, which columns are nullable, and the NULL sentinel. db-gen reads that from `information_schema.columns` and hands a template everything it needs to write it once, per language, and regenerate when the table changes.

## Configuration

```json
{
  "GenerateCopyTargets": true,
  "CopyTargetTemplate": "./templates/copy-pgx.gotmpl",
  "CopyTargetsFolderName": "copy",
  "CopyTargets": [
    {
      "Schema": "stage",
      "Table": "data_to_process",
      "MappedName": "StageData",
      "Format": "text",
      "NullString": ""
    }
  ],
  "ContextParameterMappings": [
    { "ParameterNames": ["created_by"], "ContextPath": "ctx.CreatedBy" },
    { "ParameterNames": ["job_run_id"], "ContextPath": "ctx.JobRunId" }
  ]
}
```

See the [configuration reference](./configuration.md#copy-targets) for each key. Files are written to `OutputFolder/CopyTargetsFolderName` (default `copy/`), named after each target's `MappedName` (or a generated name), using `GeneratedFileExtension` and `GeneratedFileCase`.

Context columns are matched the same way as routine context parameters — by **db column name** (case-insensitive) against `ContextParameterMappings`.

## The metadata contract

A template receives `CopyTargetTemplateData`:

```go
type CopyTargetTemplateData struct {
    Config     *Config
    CopyTarget CopyTarget
    BuildInfo  *version.BuildInformation
}

type CopyTarget struct {
    Schema          string
    Table           string
    DbFullTableName string // "schema.table"
    StructName      string // generated type/file name (or MappedName)

    Format     string // "csv" | "text" | "binary" — a hint; the template decides
    NullString string // NULL sentinel for text/CSV formats

    ContextColumns []Property // injected from context (created_by, job_run_id, …)
    DataColumns    []Property // real data columns, in ordinal order
    AllColumns     []Property // ContextColumns ++ DataColumns = the COPY column list, in wire order
}
```

Each column is a [`Property`](./templating.md#template-data). The fields that matter for copy code:

| Field | Meaning |
|-------|---------|
| `DbColumnName` | The column name — the `COPY` contract speaks column names. |
| `DbColumnType` | PostgreSQL `udt_name`, for choosing a binary type or a cast. |
| `PropertyType` | Mapped target-language type (nullable-aware) — useful for context-column parameters. |
| `Nullable` | Whether the column accepts NULL — drives the NULL-vs-empty decision. |
| `Position` | For **data** columns only: the 0-based source field index, in table-ordinal order. |
| `IsContextParameter` / `ContextPath` | Set on context columns; `ContextPath` is where the value comes from. |

### Wire order rules

- `AllColumns` is the exact `COPY (col, col, …)` list: **context columns first, then data columns**, each in order. Emit it verbatim.
- `DataColumns[i].Position == i` — data rows are expected as fields in table-ordinal order, so the template indexes the incoming row by `Position`.
- Context values are supplied once (as parameters) and injected into every row; data values stream per row.

## Example templates

Two reference templates ship under `test/templates/`. Both take data fields as text in ordinal order and turn empty nullable fields into NULL — adjust to your row source and typing.

### Go / pgx (`copy-pgx.gotmpl`)

Generates a `CopyFrom`-based loader with a `CopyFromSource` adapter. Context columns become function parameters; `AllColumns` becomes the pgx column list; data values are pulled by `Position`:

```gotemplate
var {{.CopyTarget.StructName}}Columns = []string{
{{- range .CopyTarget.AllColumns}}
	"{{.DbColumnName}}", // {{.DbColumnType}}{{if .IsContextParameter}} (context: {{.ContextPath}}){{end}}
{{- end}}
}
// …
func (s *…CopySource) Values() ([]any, error) {
	return []any{
{{- range .CopyTarget.ContextColumns}}
		s.{{camelCased .DbColumnName}}, // {{.DbColumnName}}
{{- end}}
{{- range .CopyTarget.DataColumns}}
		s.val({{.Position}}, {{.Nullable}}), // {{.DbColumnName}} {{.DbColumnType}}
{{- end}}
	}, nil
}
```

### C# / Npgsql (`copy-npgsql.gotmpl`)

Generates a binary-importer (`BeginBinaryImportAsync`) loader. `AllColumns` builds the `COPY (...)` SQL; context columns are method parameters written once per row; data fields are written by `Position` with an empty-nullable→NULL helper:

```gotemplate
private const string CopySql =
    "COPY {{.CopyTarget.DbFullTableName}} (" +
{{- range $i, $col := .CopyTarget.AllColumns}}
    "{{if $i}}, {{end}}{{$col.DbColumnName}}" +
{{- end}}
    ") FROM STDIN (FORMAT BINARY)";
```

The same metadata drives an Elixir/text-CSV template (escaping + `NULL ''`) just as well — the tool doesn't care which.

## Limitations

- Copy targets read column metadata from `information_schema` and require a live database connection (they are not part of the offline routines file).
- The example templates pass data as text and cast/stage server-side. For a strongly-typed staging table, switch the per-column writes to use `DbColumnType`.
