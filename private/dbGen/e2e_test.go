package dbGen

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update e2e golden files")

// TestE2E_Generate runs the full pipeline against the fixture database
// (GetRoutines -> Preprocess -> Process -> Generate, plus copy targets) and
// compares the generated tree against committed golden files.
//
// Regenerate goldens with:  go test ./private/dbGen -run E2E -update
func TestE2E_Generate(t *testing.T) {
	connStr := requireTestDB(t)
	outDir := t.TempDir()
	config := buildE2EConfig(connStr, outDir)

	routines, err := GetRoutines(config)
	if err != nil {
		t.Fatalf("GetRoutines: %v", err)
	}
	if err := PreprocessRoutines(&routines, config); err != nil {
		t.Fatalf("PreprocessRoutines: %v", err)
	}
	processed, err := Process(routines, config)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	copyTables, err := GetCopyTargets(config)
	if err != nil {
		t.Fatalf("GetCopyTargets: %v", err)
	}
	copyTargets, err := MapCopyTargets(copyTables, config)
	if err != nil {
		t.Fatalf("MapCopyTargets: %v", err)
	}
	if err := Generate(processed, copyTargets, config); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	goldenDir := filepath.Join("..", "..", "test", "e2e", "golden")
	compareTreeToGolden(t, outDir, goldenDir)
}

func buildE2EConfig(connStr, outDir string) *Config {
	tmpl := func(name string) string {
		return filepath.Join("..", "..", "test", "e2e", "templates", name)
	}
	return &Config{
		ConnectionString:       connStr,
		OutputFolder:           outDir,
		ModelsFolderName:       "models",
		GeneratedFileExtension: ".txt",
		GeneratedFileCase:      "pascalcase",
		GenerateModels:         true,
		GenerateProcessors:     false,
		ClearOutputFolder:      true,
		DbContextTemplate:      tmpl("dbcontext.gotmpl"),
		ModelTemplate:          tmpl("model.gotmpl"),
		Generate: []SchemaConfig{{
			Schema:       "dbgen_test",
			AllFunctions: false,
			Functions: map[string]RoutineMapping{
				"scalar_sum":   {Generate: true},
				"get_rows":     {Generate: true},
				"with_context": {Generate: true},
				// register_user drives the parameter-security coverage; the
				// per-function override on `email` (omit) must beat the default (secure).
				"register_user": {
					Generate:   true,
					Parameters: map[string]ParamMapping{"email": {SecurityLevel: "omit"}},
				},
			},
		}},
		ContextParameterMappings: []ContextParameterMapping{
			{ParameterNames: []string{"_user_id"}, ContextPath: "ctx.UserId"},
			// copy-target context columns: injected per row rather than read from the source
			{ParameterNames: []string{"created_by"}, ContextPath: "ctx.CreatedBy"},
			{ParameterNames: []string{"job_run_id"}, ContextPath: "ctx.JobRunId"},
		},
		// Parameter security: global-by-name defaults, resolved against the db
		// parameter name. `display_name` matches nothing -> DefaultParameterSecurityLevel.
		DefaultParameterSecurityLevel: "secure",
		ParameterSecurityMappings: []ParameterSecurityMapping{
			{ParameterNames: []string{"username"}, SecurityLevel: "none"},
			{ParameterNames: []string{"password", "api_token"}, SecurityLevel: "strict"},
			{ParameterNames: []string{"secret_note"}, SecurityLevel: "omit"},
		},
		// Minimal validation config so ValidationRules is non-empty for at least
		// one parameter, locking that field in the contract dump too.
		Validation: ValidationConfig{
			ValidationRuleDefinitions:   []ValidationRuleDefinition{{Name: "NotEmpty", Type: "Built-in"}},
			ParameterValidationMappings: []ParameterValidationMapping{{ParameterNames: []string{"username"}, Rules: []interface{}{"NotEmpty"}}},
		},
		AdditionalGenerators: []AdditionalGenerator{{
			Name:           "Metadata",
			Enabled:        true,
			Template:       tmpl("metadata.gotmpl"),
			OutputFolder:   filepath.Join(outDir, "metadata"),
			FileName:       "contract.txt",
			GenerationType: "single-file",
		}},
		GenerateCopyTargets:   true,
		CopyTargetTemplate:    filepath.Join("..", "..", "test", "templates", "copy-pgx.gotmpl"),
		CopyTargetsFolderName: "copy",
		CopyTargets:           []CopyTargetConfig{{Schema: "dbgen_test", Table: "copy_target_demo", Format: "csv"}},
		Mappings: []Mapping{
			{DatabaseTypes: []string{"text"}, MappedType: "string", NullableReturnType: "string?"},
			{DatabaseTypes: []string{"int4", "integer"}, MappedType: "int", NullableReturnType: "int?"},
			{DatabaseTypes: []string{"int8", "bigint"}, MappedType: "int64"},
			{DatabaseTypes: []string{"*"}, MappedType: "string"},
		},
	}
}

// compareTreeToGolden compares (or, with -update, refreshes) every file under
// genDir against goldenDir. Line endings are normalized so git autocrlf can't
// cause spurious diffs.
func compareTreeToGolden(t *testing.T, genDir, goldenDir string) {
	t.Helper()

	if *updateGolden {
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatalf("clearing golden dir: %v", err)
		}
	}

	generated := map[string]bool{}
	walk(t, genDir, func(rel string, data []byte) {
		generated[rel] = true
		goldenPath := filepath.Join(goldenDir, filepath.FromSlash(rel))

		if *updateGolden {
			if err := os.MkdirAll(filepath.Dir(goldenPath), 0o777); err != nil {
				t.Fatalf("mkdir golden: %v", err)
			}
			if err := os.WriteFile(goldenPath, data, 0o600); err != nil {
				t.Fatalf("writing golden %s: %v", rel, err)
			}
			return
		}

		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Errorf("missing golden for %s (run with -update): %v", rel, err)
			return
		}
		if !bytes.Equal(normalize(want), normalize(data)) {
			t.Errorf("generated %s differs from golden:\n--- want ---\n%s\n--- got ---\n%s", rel, want, data)
		}
	})

	if *updateGolden {
		t.Logf("golden files updated under %s", goldenDir)
		return
	}

	// catch stale goldens with no generated counterpart
	walk(t, goldenDir, func(rel string, _ []byte) {
		if !generated[rel] {
			t.Errorf("stale golden file with no generated counterpart: %s", rel)
		}
	})
}

func walk(t *testing.T, root string, fn func(rel string, data []byte)) {
	t.Helper()
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fn(filepath.ToSlash(rel), data)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
}

func normalize(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}
