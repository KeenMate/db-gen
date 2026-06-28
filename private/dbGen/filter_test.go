package dbGen

import "testing"

func TestFunctionShouldBeGenerated(t *testing.T) {
	schemaConfig := &SchemaConfig{
		AllFunctions: true,
		Functions: map[string]RoutineMapping{
			"ignored":          {Generate: false},
			"explicit_include": {Generate: true},
		},
	}

	cases := map[string]bool{
		"ignored":          false, // explicitly disabled
		"explicit_include": true,  // explicitly enabled
		"some_other":       true,  // not listed -> AllFunctions
	}
	for name, want := range cases {
		if got := functionShouldBeGenerated(name, schemaConfig); got != want {
			t.Errorf("functionShouldBeGenerated(%q) = %v, want %v", name, got, want)
		}
	}

	// AllFunctions false: only explicit ones generate
	deny := &SchemaConfig{AllFunctions: false, Functions: map[string]RoutineMapping{"keep": {Generate: true}}}
	if functionShouldBeGenerated("not_listed", deny) {
		t.Error("expected not_listed to be excluded when AllFunctions is false")
	}
	if !functionShouldBeGenerated("keep", deny) {
		t.Error("expected explicit keep to be included")
	}
}

func TestFilterFunctions(t *testing.T) {
	routines := []DbRoutine{
		{RoutineSchema: "public", RoutineName: "a"},
		{RoutineSchema: "public", RoutineName: "ignored"},
		{RoutineSchema: "other", RoutineName: "b"},      // schema not configured -> dropped
		{RoutineSchema: "app", RoutineName: "included"}, // explicit include in non-all schema
	}
	config := &Config{
		Generate: []SchemaConfig{
			{Schema: "public", AllFunctions: true, Functions: map[string]RoutineMapping{"ignored": {Generate: false}}},
			{Schema: "app", AllFunctions: false, Functions: map[string]RoutineMapping{"included": {Generate: true}}},
		},
	}

	filtered, err := FilterFunctions(&routines, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := map[string]bool{}
	for _, r := range filtered {
		got[r.RoutineSchema+"."+r.RoutineName] = true
	}
	want := []string{"public.a", "app.included"}
	if len(filtered) != len(want) {
		t.Fatalf("filtered = %v, want %v", got, want)
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("expected %q in filtered set %v", w, got)
		}
	}
}
