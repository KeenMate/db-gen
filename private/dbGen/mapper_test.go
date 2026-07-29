package dbGen

import "testing"

func TestGetTypeMapping(t *testing.T) {
	mappings := map[string]mapping{
		"int4": {mappedType: "int", mappedFunction: "GetInt32"},
		"*":    {mappedType: "string", mappedFunction: "GetString"},
	}

	t.Run("known type", func(t *testing.T) {
		m, err := getTypeMapping("int4", &mappings)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.mappedType != "int" {
			t.Errorf("mappedType = %q, want int", m.mappedType)
		}
	})

	t.Run("falls back to *", func(t *testing.T) {
		m, err := getTypeMapping("citext", &mappings)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.mappedType != "string" {
			t.Errorf("fallback mappedType = %q, want string", m.mappedType)
		}
	})

	t.Run("no mapping, no fallback -> error", func(t *testing.T) {
		noFallback := map[string]mapping{"int4": {mappedType: "int"}}
		if _, err := getTypeMapping("uuid", &noFallback); err == nil {
			t.Error("expected error for unmapped type without fallback")
		}
	})
}

func TestGetFunctionName(t *testing.T) {
	cases := []struct {
		db, schema, mapped, want string
	}{
		{"get_user", "public", "", "GetUser"},      // public schema is hidden
		{"get_user", "app", "", "AppGetUser"},      // non-public prefixes schema
		{"get_user", "public", "Custom", "Custom"}, // explicit mapped name wins
		{"get_user", "app", "Custom", "Custom"},    // mapped name wins over schema
	}
	for _, c := range cases {
		if got := getFunctionName(c.db, c.schema, c.mapped); got != c.want {
			t.Errorf("getFunctionName(%q,%q,%q) = %q, want %q", c.db, c.schema, c.mapped, got, c.want)
		}
	}
}

func TestModelAndProcessorName(t *testing.T) {
	if got := getModelName("GetUser"); got != "GetUserModel" {
		t.Errorf("getModelName = %q, want GetUserModel", got)
	}
	if got := getProcessorName("GetUser"); got != "GetUserProcessor" {
		t.Errorf("getProcessorName = %q, want GetUserProcessor", got)
	}
}

func TestProcessContextParameters(t *testing.T) {
	config := &Config{
		ContextParameterMappings: []ContextParameterMapping{
			{ParameterNames: []string{"_user_id"}, ContextPath: "ctx.UserId"},
		},
	}
	params := []Property{
		{PropertyName: "_user_id"},
		{PropertyName: "name"},
	}

	ctxParams, regular, all := processContextParameters(params, config)

	if len(ctxParams) != 1 || ctxParams[0].PropertyName != "_user_id" {
		t.Fatalf("context params = %+v, want one _user_id", ctxParams)
	}
	if !ctxParams[0].IsContextParameter || ctxParams[0].ContextPath != "ctx.UserId" {
		t.Errorf("context param not marked correctly: %+v", ctxParams[0])
	}
	if len(regular) != 1 || regular[0].PropertyName != "name" {
		t.Fatalf("regular params = %+v, want one name", regular)
	}
	if len(all) != 2 || all[0].PropertyName != "_user_id" || all[1].PropertyName != "name" {
		t.Errorf("all params order wrong: %+v", all)
	}
}

func TestMapParameters_TypeResolution(t *testing.T) {
	attrs := []DbParameter{
		{OrdinalPosition: 1, Name: "a", Mode: InMode, UDTName: "int4"},
		{OrdinalPosition: 2, Name: "b", Mode: InMode, UDTName: "int4", IsOptional: true},
		{OrdinalPosition: 3, Name: "c", Mode: InMode, UDTName: "int4", IsNullable: true},
	}
	mappings := map[string]mapping{
		"int4": {mappedType: "int", nullableParameterType: "int?", optionalParameterType: "Optional<int>"},
	}
	rm := emptyMapping
	props, err := mapParameters(attrs, &mappings, &rm, &Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 3 {
		t.Fatalf("got %d props, want 3", len(props))
	}
	if props[0].PropertyType != "int" {
		t.Errorf("plain param type = %q, want int", props[0].PropertyType)
	}
	if props[1].PropertyType != "Optional<int>" {
		t.Errorf("optional param type = %q, want Optional<int>", props[1].PropertyType)
	}
	if props[2].PropertyType != "int?" {
		t.Errorf("nullable param type = %q, want int?", props[2].PropertyType)
	}
	// positions are zero-based relative to the first ordinal
	if props[0].Position != 0 || props[2].Position != 2 {
		t.Errorf("positions wrong: %d, %d", props[0].Position, props[2].Position)
	}
}

func TestResolveSecurityLevel(t *testing.T) {
	config := &Config{
		DefaultParameterSecurityLevel: SecurityLevelSecure,
		ParameterSecurityMappings: []ParameterSecurityMapping{
			{ParameterNames: []string{"password", "secret"}, SecurityLevel: SecurityLevelStrict},
			{ParameterNames: []string{"email"}, SecurityLevel: SecurityLevelNone},
		},
	}

	t.Run("per-function override wins over global and default", func(t *testing.T) {
		got := resolveSecurityLevel("password", ParamMapping{SecurityLevel: SecurityLevelOmit}, true, config)
		if got != SecurityLevelOmit {
			t.Errorf("got %q, want %q", got, SecurityLevelOmit)
		}
	})

	t.Run("global by-name match, case-insensitive", func(t *testing.T) {
		got := resolveSecurityLevel("PASSWORD", ParamMapping{}, false, config)
		if got != SecurityLevelStrict {
			t.Errorf("got %q, want %q", got, SecurityLevelStrict)
		}
	})

	t.Run("empty per-function level falls through to global", func(t *testing.T) {
		got := resolveSecurityLevel("secret", ParamMapping{SecurityLevel: ""}, true, config)
		if got != SecurityLevelStrict {
			t.Errorf("got %q, want %q", got, SecurityLevelStrict)
		}
	})

	t.Run("no match falls back to default", func(t *testing.T) {
		got := resolveSecurityLevel("username", ParamMapping{}, false, config)
		if got != SecurityLevelSecure {
			t.Errorf("got %q, want %q", got, SecurityLevelSecure)
		}
	})
}

func TestNormalizeAndValidateSecurityLevels(t *testing.T) {
	t.Run("empty default becomes DefaultSecurityLevel; values lowercased", func(t *testing.T) {
		config := &Config{
			ParameterSecurityMappings: []ParameterSecurityMapping{
				{ParameterNames: []string{"password"}, SecurityLevel: "STRICT"},
			},
			Generate: []SchemaConfig{{
				Schema: "public",
				Functions: map[string]RoutineMapping{
					"create_user": {Parameters: map[string]ParamMapping{
						"pwd": {SecurityLevel: "Omit"},
					}},
				},
			}},
		}
		if err := normalizeAndValidateSecurityLevels(config); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config.DefaultParameterSecurityLevel != DefaultSecurityLevel {
			t.Errorf("default = %q, want %q", config.DefaultParameterSecurityLevel, DefaultSecurityLevel)
		}
		if config.ParameterSecurityMappings[0].SecurityLevel != SecurityLevelStrict {
			t.Errorf("global level not normalized: %q", config.ParameterSecurityMappings[0].SecurityLevel)
		}
		if got := config.Generate[0].Functions["create_user"].Parameters["pwd"].SecurityLevel; got != SecurityLevelOmit {
			t.Errorf("per-function level not normalized: %q", got)
		}
	})

	t.Run("invalid value is rejected", func(t *testing.T) {
		config := &Config{DefaultParameterSecurityLevel: "loud"}
		if err := normalizeAndValidateSecurityLevels(config); err == nil {
			t.Error("expected error for invalid security level")
		}
	})
}

func TestMapModel_Scalar(t *testing.T) {
	routine := DbRoutine{RoutineName: "sum", RoutineSchema: "public", DataType: "int4", FuncType: "function"}
	mappings := map[string]mapping{"int4": {mappedType: "int", mappedFunction: "GetInt32"}}
	rm := emptyMapping
	props, err := mapModel(routine, &mappings, &rm, &Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("scalar model should have 1 property, got %d", len(props))
	}
	if props[0].PropertyName != "Sum" || props[0].PropertyType != "int" || props[0].MapperFunction != "GetInt32" {
		t.Errorf("scalar property wrong: %+v", props[0])
	}
}

func TestMapModel_StructuredNullableReturn(t *testing.T) {
	routine := DbRoutine{
		RoutineName: "get_row", RoutineSchema: "public", DataType: "record", FuncType: "function",
		OutParameters: []DbParameter{
			{OrdinalPosition: 1, Name: "id", Mode: OutMode, UDTName: "int4"},
			{OrdinalPosition: 2, Name: "label", Mode: OutMode, UDTName: "text", IsNullable: true},
		},
	}
	mappings := map[string]mapping{
		"int4": {mappedType: "int"},
		"text": {mappedType: "string", nullableReturnType: "string?"},
	}
	rm := emptyMapping
	props, err := mapModel(routine, &mappings, &rm, &Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("got %d props, want 2", len(props))
	}
	if props[0].PropertyType != "int" {
		t.Errorf("id type = %q, want int", props[0].PropertyType)
	}
	if props[1].PropertyType != "string?" {
		t.Errorf("nullable label type = %q, want string?", props[1].PropertyType)
	}
}
