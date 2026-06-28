package dbGen

import (
	"strings"
	"testing"
)

func TestGetParameterChanges(t *testing.T) {
	old := []DbParameter{
		{OrdinalPosition: 1, Name: "a", UDTName: "int4"},
		{OrdinalPosition: 2, Name: "b", UDTName: "int4", IsNullable: false},
		{OrdinalPosition: 3, Name: "to_remove", UDTName: "text"},
	}
	newP := []DbParameter{
		{OrdinalPosition: 1, Name: "a_renamed", UDTName: "int4"}, // renamed
		{OrdinalPosition: 2, Name: "b", UDTName: "int8"},         // type changed
	}

	out := getParameterChanges(old, newP)
	wantSubstrings := []string{"renamed from a", "data type changed from int4 to int8", "removed parameter: to_remove"}
	for _, w := range wantSubstrings {
		if !strings.Contains(out, w) {
			t.Errorf("parameter changes missing %q\ngot:\n%s", w, out)
		}
	}
}

func TestGetColumnChanges(t *testing.T) {
	old := []DbColumn{
		{OrdinalPosition: 1, Name: "id", UDTName: "int4", IsNullable: false},
		{OrdinalPosition: 2, Name: "city", UDTName: "text", IsNullable: true},
	}
	newC := []DbColumn{
		{OrdinalPosition: 1, Name: "id", UDTName: "int8", IsNullable: false},    // type changed
		{OrdinalPosition: 2, Name: "city", UDTName: "text", IsNullable: false},  // nullability changed
		{OrdinalPosition: 3, Name: "region", UDTName: "text", IsNullable: true}, // added
	}

	out := getColumnChanges(old, newC)
	wantSubstrings := []string{
		"data type changed from int4 to int8",
		"nullability changed from true to false",
		"added column: region",
	}
	for _, w := range wantSubstrings {
		if !strings.Contains(out, w) {
			t.Errorf("column changes missing %q\ngot:\n%s", w, out)
		}
	}
}

func TestGetRoutinesChanges(t *testing.T) {
	info := &GenerationInformation{
		Routines: []DbRoutine{
			{RoutineSchema: "public", RoutineName: "deleted_fn", RoutineNameWithParams: "deleted_fn()"},
			{RoutineSchema: "public", RoutineName: "kept_fn", RoutineNameWithParams: "kept_fn()"},
		},
	}
	newRoutines := []DbRoutine{
		{RoutineSchema: "public", RoutineName: "kept_fn", RoutineNameWithParams: "kept_fn()"},
		{RoutineSchema: "public", RoutineName: "new_fn", RoutineNameWithParams: "new_fn()"},
	}

	out := info.GetRoutinesChanges(newRoutines)
	if !strings.Contains(out, "Deleted routines:") || !strings.Contains(out, "deleted_fn") {
		t.Errorf("expected deleted_fn reported\ngot:\n%s", out)
	}
	if !strings.Contains(out, "Created routines:") || !strings.Contains(out, "new_fn") {
		t.Errorf("expected new_fn reported\ngot:\n%s", out)
	}
}

func TestGetTableChanges(t *testing.T) {
	info := &GenerationInformation{
		Tables: []DbTable{
			{TableSchema: "stage", TableName: "kept", Columns: []DbColumn{
				{OrdinalPosition: 1, Name: "id", UDTName: "int4"},
			}},
			{TableSchema: "stage", TableName: "gone", Columns: []DbColumn{
				{OrdinalPosition: 1, Name: "id", UDTName: "int4"},
			}},
		},
	}
	newTables := []DbTable{
		{TableSchema: "stage", TableName: "kept", Columns: []DbColumn{
			{OrdinalPosition: 1, Name: "id", UDTName: "int4"},
			{OrdinalPosition: 2, Name: "added_col", UDTName: "text"},
		}},
		{TableSchema: "stage", TableName: "fresh", Columns: []DbColumn{
			{OrdinalPosition: 1, Name: "id", UDTName: "int4"},
		}},
	}

	out := info.GetTableChanges(newTables)
	wantSubstrings := []string{
		"Deleted copy-target tables:",
		"stage.gone",
		"Created copy-target tables:",
		"stage.fresh",
		"Changed copy-target tables:",
		"added column: added_col",
	}
	for _, w := range wantSubstrings {
		if !strings.Contains(out, w) {
			t.Errorf("table changes missing %q\ngot:\n%s", w, out)
		}
	}
}

func TestGetTableChanges_NoChanges(t *testing.T) {
	cols := []DbColumn{{OrdinalPosition: 1, Name: "id", UDTName: "int4"}}
	info := &GenerationInformation{Tables: []DbTable{{TableSchema: "stage", TableName: "t", Columns: cols}}}
	newTables := []DbTable{{TableSchema: "stage", TableName: "t", Columns: cols}}

	if out := info.GetTableChanges(newTables); out != "" {
		t.Errorf("expected no changes, got:\n%s", out)
	}
}
