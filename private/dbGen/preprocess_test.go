package dbGen

import "testing"

func TestGetTypeMappings(t *testing.T) {
	config := &Config{
		Mappings: []Mapping{
			{DatabaseTypes: []string{"integer", "int4"}, MappedType: "int", MappingFunction: "GetInt32"},
			{DatabaseTypes: []string{"text"}, MappedType: "string", NullableReturnType: "string?"},
		},
	}
	m := getTypeMappings(config)

	if m["integer"].mappedType != "int" || m["int4"].mappedType != "int" {
		t.Errorf("integer/int4 not both mapped to int: %+v", m)
	}
	if m["int4"].mappedFunction != "GetInt32" {
		t.Errorf("mappedFunction not carried: %+v", m["int4"])
	}
	if m["text"].nullableReturnType != "string?" {
		t.Errorf("nullableReturnType not carried: %+v", m["text"])
	}
}

func TestMarkOverloadedRoutines(t *testing.T) {
	routines := []DbRoutine{
		{RoutineSchema: "public", RoutineName: "f", RoutineNameWithParams: "f(int4)"},
		{RoutineSchema: "public", RoutineName: "f", RoutineNameWithParams: "f(text)"},
		{RoutineSchema: "public", RoutineName: "g", RoutineNameWithParams: "g()"},
	}
	config := &Config{Generate: []SchemaConfig{{Schema: "public", AllFunctions: true}}}

	if err := PreprocessRoutines(&routines, config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !routines[0].HasOverload || !routines[1].HasOverload {
		t.Errorf("duplicate-named routines should be marked overloaded: %+v", routines)
	}
	if routines[2].HasOverload {
		t.Errorf("unique routine g should not be overloaded")
	}
}
