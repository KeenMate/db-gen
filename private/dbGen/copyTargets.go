package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/private/database"
	common2 "github.com/keenmate/db-gen/private/helpers"
	"github.com/keenmate/db-gen/private/version"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Data source: tables/columns (parallel to routinesProvider.go for routines)
// ---------------------------------------------------------------------------

// DbTable is the raw shape read from information_schema.columns.
type DbTable struct {
	TableSchema string `db:"table_schema"`
	TableName   string `db:"table_name"`
	Columns     []DbColumn
}

// DbColumn mirrors DbParameter but for a real table column.
type DbColumn struct {
	OrdinalPosition int    `db:"ordinal_position"`
	Name            string `db:"column_name"`
	UDTName         string `db:"udt_name"`
	IsNullable      bool   `db:"is_nullable"`
	HasDefault      bool   `db:"has_default"`
}

// ---------------------------------------------------------------------------
// Template data: the metadata a per-language COPY template consumes.
// A column is already modelled by Property, so we reuse it verbatim. The split
// into ContextColumns/DataColumns mirrors Routine.ContextParameters/RegularParameters.
// ---------------------------------------------------------------------------

type CopyTarget struct {
	Schema          string
	Table           string
	DbFullTableName string // schema.table
	StructName      string // generated type/name, e.g. StageResCompany

	// Wire-format hints for the template. db-gen stays language-agnostic; the
	// template decides what to do with these.
	Format     string // "csv" | "text" | "binary" (purely a hint for the template)
	NullString string // NULL sentinel for text/CSV formats (e.g. "")

	ContextColumns []Property // injected from context (created_by, job_run_id, ...)
	DataColumns    []Property // real data columns, in ordinal order (Position = source field index)
	AllColumns     []Property // ContextColumns ++ DataColumns = the COPY column list, in wire order
}

type CopyTargetTemplateData struct {
	Config     *Config
	CopyTarget CopyTarget
	BuildInfo  *version.BuildInformation
}

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

// GetCopyTargets reads the configured tables' column metadata from the database.
func GetCopyTargets(config *Config) ([]DbTable, error) {
	if len(config.CopyTargets) == 0 {
		return nil, nil
	}

	log.Printf("Connecting to database (copy targets)...")
	conn, err := database.Connect(config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %s", err)
	}

	tables := make([]DbTable, 0, len(config.CopyTargets))
	for _, ct := range config.CopyTargets {
		columns, err := getColumnsForTable(conn, ct.Schema, ct.Table)
		if err != nil {
			return nil, fmt.Errorf("getting columns for %s.%s: %s", ct.Schema, ct.Table, err)
		}

		if len(columns) == 0 {
			return nil, fmt.Errorf("table %s.%s has no columns (does it exist?)", ct.Schema, ct.Table)
		}

		tables = append(tables, DbTable{
			TableSchema: ct.Schema,
			TableName:   ct.Table,
			Columns:     columns,
		})
	}

	return tables, nil
}

func getColumnsForTable(conn *database.DbConn, schema string, table string) ([]DbColumn, error) {
	columns := new([]DbColumn)

	q := `
		select ordinal_position::int,
		       column_name::text,
		       udt_name::text,
		       is_nullable = 'YES' as is_nullable,
		       column_default is not null as has_default
		from information_schema.columns
		where table_schema = $1
		  and table_name = $2
		order by ordinal_position;`

	err := conn.Select(columns, q, schema, table)
	if err != nil {
		return nil, err
	}

	return *columns, nil
}

// ---------------------------------------------------------------------------
// Mapping (parallel to mapRoutines)
// ---------------------------------------------------------------------------

// MapCopyTargets maps raw table metadata into template-ready copy targets.
func MapCopyTargets(tables []DbTable, config *Config) ([]CopyTarget, error) {
	if !config.GenerateCopyTargets || len(tables) == 0 {
		return nil, nil
	}

	typeMappings := getTypeMappings(config)
	return mapCopyTargets(tables, &typeMappings, config)
}

func mapCopyTargets(tables []DbTable, globalTypeMappings *map[string]mapping, config *Config) ([]CopyTarget, error) {
	cfgByName := make(map[string]CopyTargetConfig)
	for _, ct := range config.CopyTargets {
		cfgByName[ct.Schema+"."+ct.Table] = ct
	}

	targets := make([]CopyTarget, 0, len(tables))
	for _, table := range tables {
		ctCfg := cfgByName[table.TableSchema+"."+table.TableName]

		columns, err := mapColumns(table, globalTypeMappings)
		if err != nil {
			return nil, err
		}

		// Split into context-injected vs data columns by db column name.
		contextCols, dataCols := splitCopyColumns(columns, config)

		// Source field index: data columns are expected in ordinal order.
		for i := range dataCols {
			dataCols[i].Position = i
		}

		// COPY column list / wire order: context columns first, then data columns.
		allCols := make([]Property, 0, len(columns))
		allCols = append(allCols, contextCols...)
		allCols = append(allCols, dataCols...)

		name := ctCfg.MappedName
		if name == "" {
			name = getCopyTargetName(table.TableSchema, table.TableName)
		}

		format := ctCfg.Format
		if format == "" {
			format = "csv"
		}

		targets = append(targets, CopyTarget{
			Schema:          table.TableSchema,
			Table:           table.TableName,
			DbFullTableName: table.TableSchema + "." + table.TableName,
			StructName:      name,
			Format:          format,
			NullString:      ctCfg.NullString,
			ContextColumns:  contextCols,
			DataColumns:     dataCols,
			AllColumns:      allCols,
		})
	}

	return targets, nil
}

func mapColumns(table DbTable, globalTypeMappings *map[string]mapping) ([]Property, error) {
	cols := append([]DbColumn(nil), table.Columns...)
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].OrdinalPosition < cols[j].OrdinalPosition
	})

	properties := make([]Property, 0, len(cols))
	for _, c := range cols {
		typeMapping, err := getTypeMapping(c.UDTName, globalTypeMappings)
		if err != nil {
			return nil, fmt.Errorf("mapping column %s.%s.%s: %s", table.TableSchema, table.TableName, c.Name, err)
		}

		propertyType := typeMapping.mappedType
		if c.IsNullable && typeMapping.nullableReturnType != "" {
			propertyType = typeMapping.nullableReturnType
		}

		properties = append(properties, Property{
			DbColumnName:       c.Name,
			DbColumnType:       c.UDTName,
			PropertyName:       common2.ToPascalCase(c.Name),
			PropertyType:       propertyType,
			BaseType:           typeMapping.mappedType,
			NullableReturnType: typeMapping.nullableReturnType,
			MapperFunction:     typeMapping.mappedFunction,
			Nullable:           c.IsNullable,
		})
	}

	return properties, nil
}

// splitCopyColumns is the copy-target equivalent of processContextParameters,
// but matches on the raw db column name (the COPY contract speaks column names).
func splitCopyColumns(columns []Property, config *Config) (contextCols []Property, dataCols []Property) {
	contextMap := make(map[string]string)
	for _, m := range config.ContextParameterMappings {
		for _, paramName := range m.ParameterNames {
			contextMap[strings.ToLower(paramName)] = m.ContextPath
		}
	}

	for _, col := range columns {
		if contextPath, isContext := contextMap[strings.ToLower(col.DbColumnName)]; isContext {
			col.IsContextParameter = true
			col.ContextPath = contextPath
			contextCols = append(contextCols, col)
		} else {
			dataCols = append(dataCols, col)
		}
	}

	return contextCols, dataCols
}

func getCopyTargetName(schema string, table string) string {
	prefix := ""
	if schema != hiddenSchema {
		prefix = common2.ToPascalCase(common2.NormalizeStr(schema))
	}
	return prefix + common2.ToPascalCase(table)
}

// ---------------------------------------------------------------------------
// Generation (parallel to generatePerRoutineFiles)
// ---------------------------------------------------------------------------

func generateCopyTargets(targets []CopyTarget, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	if !config.GenerateCopyTargets || len(targets) == 0 {
		return nil
	}

	log.Printf("Generating copy targets...")

	tmpl, err := parseTemplate(config.CopyTargetTemplate)
	if err != nil {
		return fmt.Errorf("loading copy target template: %s", err)
	}

	outFolder := filepath.Join(config.OutputFolder, config.CopyTargetsFolderName)
	err = os.MkdirAll(outFolder, 0777)
	if err != nil {
		return fmt.Errorf("creating copy targets output folder: %s", err)
	}

	for _, target := range targets {
		filename := changeCase(target.StructName+config.GeneratedFileExtension, config.GeneratedFileCase)
		relPath := filepath.Join(config.CopyTargetsFolderName, filename)
		filePath := filepath.Join(config.OutputFolder, relPath)

		data := &CopyTargetTemplateData{
			Config:     config,
			CopyTarget: target,
			BuildInfo:  version.GetBuildInfo(),
		}

		changed, err := generateFile(data, tmpl, filePath, hashMap, generatedFiles)
		if err != nil {
			return fmt.Errorf("generating copy target %s: %s", target.StructName, err)
		}

		if changed {
			log.Printf("Updated: %s", relPath)
		} else {
			common2.LogDebug("Same: %s", relPath)
		}
	}

	return nil
}
