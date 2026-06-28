package dbGen

import "testing"

func TestGetRoutines_Integration(t *testing.T) {
	connStr := requireTestDB(t)
	config := &Config{
		ConnectionString: connStr,
		Generate:         []SchemaConfig{{Schema: "dbgen_test", AllFunctions: true}},
	}

	routines, err := GetRoutines(config)
	if err != nil {
		t.Fatalf("GetRoutines: %v", err)
	}

	byName := map[string]DbRoutine{}
	overloadCount := 0
	for _, r := range routines {
		byName[r.RoutineName] = r
		if r.RoutineName == "overloaded" {
			overloadCount++
		}
	}

	for _, want := range []string{"scalar_sum", "get_rows", "do_nothing", "noop_proc", "with_defaults", "with_context", "overloaded"} {
		if _, ok := byName[want]; !ok {
			t.Errorf("expected routine %q in %v", want, keys(byName))
		}
	}

	if byName["noop_proc"].FuncType != Procedure {
		t.Errorf("noop_proc FuncType = %q, want procedure", byName["noop_proc"].FuncType)
	}
	if overloadCount != 2 {
		t.Errorf("expected 2 overloaded functions, got %d", overloadCount)
	}

	// with_defaults has one DEFAULT (optional) parameter
	wd := byName["with_defaults"]
	optional := false
	for _, p := range wd.InParameters {
		if p.Name == "optional_val" && p.IsOptional {
			optional = true
		}
	}
	if !optional {
		t.Errorf("with_defaults.optional_val should be optional: %+v", wd.InParameters)
	}
}

func TestGetCopyTargets_Integration(t *testing.T) {
	connStr := requireTestDB(t)
	config := &Config{
		ConnectionString:    connStr,
		GenerateCopyTargets: true,
		CopyTargets:         []CopyTargetConfig{{Schema: "dbgen_test", Table: "copy_target_demo"}},
	}

	tables, err := GetCopyTargets(config)
	if err != nil {
		t.Fatalf("GetCopyTargets: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(tables))
	}

	cols := tables[0].Columns
	if len(cols) != 5 {
		t.Fatalf("got %d columns, want 5: %+v", len(cols), cols)
	}

	// columns come back in ordinal order
	wantOrder := []string{"created_by", "job_run_id", "company_id", "name", "amount"}
	for i, w := range wantOrder {
		if cols[i].Name != w {
			t.Errorf("column %d = %q, want %q", i, cols[i].Name, w)
		}
	}

	byName := map[string]DbColumn{}
	for _, c := range cols {
		byName[c.Name] = c
	}
	if byName["job_run_id"].UDTName != "int8" {
		t.Errorf("job_run_id udt = %q, want int8", byName["job_run_id"].UDTName)
	}
	if byName["company_id"].IsNullable {
		t.Error("company_id should be NOT NULL")
	}
	if !byName["name"].IsNullable {
		t.Error("name should be nullable")
	}
}

func keys(m map[string]DbRoutine) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
