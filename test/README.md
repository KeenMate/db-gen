# db-gen test framework

Tests are layered from pure/fast to DB-backed/slow. Everything runs through the
standard Go test tool.

```
make test-unit          # pure-Go unit tests, no database
make test-integration   # DB-backed integration + e2e (needs a test database)
make test               # everything (DB tests auto-skip if no DB is reachable)
make test-update-golden # regenerate e2e golden files after template changes
```

## Layers

### 1. Unit tests (no database)
Pure transformation logic, fast and deterministic. Located next to the code in
`private/`:

| File | Covers |
|------|--------|
| `private/helpers/strings_test.go` | case conversion (`ToPascalCase`/`ToCamelCase`/`ToSnakeCase`/`NormalizeStr`) |
| `private/dbGen/mapper_test.go` | type-mapping fallback, 3-tier param/model type resolution, context-param split, name generation |
| `private/dbGen/filter_test.go` | schema/function filtering |
| `private/dbGen/preprocess_test.go` | overload detection, global type-mapping table |
| `private/dbGen/changedetection_test.go` | routine diff and copy-target (table) diff |
| `private/dbGen/copytargets_test.go` | column mapping, context split, copy-target assembly |
| `private/dbGen/generator_test.go` | case-of-filename, hashing, template render + change detection, orphan removal, output-folder clearing |
| `private/dbGen/config_test.go` | path normalization, local-config discovery, full viper load |

Run only these: `go test -short ./private/...`

### 2. Integration tests (need a database)
`private/dbGen/integration_test.go` — exercises the real DB queries
(`GetRoutines`, `GetCopyTargets`) against fixtures.

### 3. End-to-end golden tests
`private/dbGen/e2e_test.go` — runs the full pipeline
(`GetRoutines → Preprocess → Process → GetCopyTargets → MapCopyTargets → Generate`)
and compares the generated tree against committed golden files in
`test/e2e/golden/`. Templates live in `test/e2e/templates/` (minimal and
deterministic) plus the shared `test/templates/copy-pgx.gotmpl`.

After an intentional change to templates or generation output:

```
make test-update-golden
git diff test/e2e/golden   # review the change
```

## Database setup

DB-backed tests need a reachable Postgres. The connection string is resolved as:

1. `DBGEN_TEST_DSN` environment variable, if set; otherwise
2. the `ConnectionString` from `test/db-gen-copy.json`.

If neither connects, DB-backed tests **skip** (they never fail the suite).

The fixtures in `test/fixtures/schema.sql` create an isolated `dbgen_test`
schema (functions covering scalar/table/void/procedure/optional/context/overload
cases, plus a copy-target staging table). The harness drops and recreates that
schema on every run, so tests always see a known structure — they do **not**
depend on the rest of the database.

The Postgres instance itself is provisioned by the sibling `db-gen-database`
project (`make setup`), which builds the `db_gen` database the test DSN points at.

## Adding tests

- **Pure logic** → add a `*_test.go` next to the code (package `dbGen`, white-box).
- **New DB object** → add it to `test/fixtures/schema.sql`, then assert against it
  in an integration test using `requireTestDB(t)`.
- **New generation feature** → extend the e2e config in `buildE2EConfig`, run
  `make test-update-golden`, and commit the new golden files.
