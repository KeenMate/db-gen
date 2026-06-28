package dbGen

import "testing"

func TestGetCopyTargetName(t *testing.T) {
	cases := []struct {
		schema, table, want string
	}{
		{"stage", "data_to_process", "StageDataToProcess"},
		{"public", "companies", "Companies"}, // public schema is hidden
	}
	for _, c := range cases {
		if got := getCopyTargetName(c.schema, c.table); got != c.want {
			t.Errorf("getCopyTargetName(%q,%q) = %q, want %q", c.schema, c.table, got, c.want)
		}
	}
}

func TestSplitCopyColumns(t *testing.T) {
	config := &Config{
		ContextParameterMappings: []ContextParameterMapping{
			{ParameterNames: []string{"created_by", "job_run_id"}, ContextPath: "ctx"},
		},
	}
	cols := []Property{
		{DbColumnName: "created_by"},
		{DbColumnName: "job_run_id"},
		{DbColumnName: "name"},
	}

	ctxCols, dataCols := splitCopyColumns(cols, config)
	if len(ctxCols) != 2 {
		t.Fatalf("context columns = %d, want 2", len(ctxCols))
	}
	if !ctxCols[0].IsContextParameter || ctxCols[0].ContextPath != "ctx" {
		t.Errorf("context column not marked: %+v", ctxCols[0])
	}
	if len(dataCols) != 1 || dataCols[0].DbColumnName != "name" {
		t.Fatalf("data columns = %+v, want [name]", dataCols)
	}
}

func TestMapColumns(t *testing.T) {
	table := DbTable{
		TableSchema: "stage", TableName: "t",
		Columns: []DbColumn{
			// intentionally out of order to verify sorting
			{OrdinalPosition: 2, Name: "name", UDTName: "text", IsNullable: true},
			{OrdinalPosition: 1, Name: "id", UDTName: "int8", IsNullable: false},
		},
	}
	mappings := map[string]mapping{
		"int8": {mappedType: "int64"},
		"text": {mappedType: "string", nullableReturnType: "string?"},
	}

	props, err := mapColumns(table, &mappings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("got %d props, want 2", len(props))
	}
	if props[0].DbColumnName != "id" || props[1].DbColumnName != "name" {
		t.Errorf("columns not sorted by ordinal: %+v", props)
	}
	if props[0].PropertyName != "Id" || props[0].PropertyType != "int64" {
		t.Errorf("id property wrong: %+v", props[0])
	}
	if props[1].PropertyType != "string?" { // nullable -> nullableReturnType
		t.Errorf("nullable name type = %q, want string?", props[1].PropertyType)
	}
}

func TestMapCopyTargets(t *testing.T) {
	tables := []DbTable{{
		TableSchema: "stage", TableName: "data_to_process",
		Columns: []DbColumn{
			{OrdinalPosition: 1, Name: "created_by", UDTName: "text"},
			{OrdinalPosition: 2, Name: "job_run_id", UDTName: "int8"},
			{OrdinalPosition: 3, Name: "company_id", UDTName: "text"},
			{OrdinalPosition: 4, Name: "name", UDTName: "text", IsNullable: true},
		},
	}}
	config := &Config{
		ContextParameterMappings: []ContextParameterMapping{
			{ParameterNames: []string{"created_by", "job_run_id"}, ContextPath: "ctx"},
		},
		CopyTargets: []CopyTargetConfig{
			{Schema: "stage", Table: "data_to_process", Format: "csv", NullString: ""},
		},
	}
	mappings := map[string]mapping{
		"text": {mappedType: "string"},
		"int8": {mappedType: "int64"},
	}

	targets, err := mapCopyTargets(tables, &mappings, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(targets))
	}
	tg := targets[0]

	if tg.StructName != "StageDataToProcess" {
		t.Errorf("StructName = %q", tg.StructName)
	}
	if tg.DbFullTableName != "stage.data_to_process" {
		t.Errorf("DbFullTableName = %q", tg.DbFullTableName)
	}
	if len(tg.ContextColumns) != 2 || len(tg.DataColumns) != 2 {
		t.Fatalf("split wrong: ctx=%d data=%d", len(tg.ContextColumns), len(tg.DataColumns))
	}
	// AllColumns wire order: context first, then data
	if len(tg.AllColumns) != 4 ||
		tg.AllColumns[0].DbColumnName != "created_by" ||
		tg.AllColumns[2].DbColumnName != "company_id" {
		t.Errorf("AllColumns order wrong: %+v", tg.AllColumns)
	}
	// data columns get sequential source positions
	if tg.DataColumns[0].Position != 0 || tg.DataColumns[1].Position != 1 {
		t.Errorf("data positions wrong: %d, %d", tg.DataColumns[0].Position, tg.DataColumns[1].Position)
	}
	if tg.DataColumns[1].PropertyType != "string" || !tg.DataColumns[1].Nullable {
		t.Errorf("nullable data column wrong: %+v", tg.DataColumns[1])
	}
}
