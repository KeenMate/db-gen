package dbGen

import (
	"strings"
	"testing"
)

// cleanSettings returns a config that should produce no validation errors.
func cleanSettings() *Config {
	return &Config{
		PathBase:               "/base",
		OutputFolder:           "/base/out",
		DbContextTemplate:      "/base/dbcontext.gotmpl",
		GeneratedFileExtension: ".cs",
		GeneratedFileCase:      "pascalcase",
		Generate:               []SchemaConfig{{Schema: "public", AllFunctions: true}},
		Mappings:               []Mapping{{DatabaseTypes: []string{"*"}, MappedType: "string"}},
	}
}

func hasIssue(issues []ValidationIssue, sev IssueSeverity, targetSubstr string) bool {
	for _, i := range issues {
		if i.Severity == sev && strings.Contains(i.Target, targetSubstr) {
			return true
		}
	}
	return false
}

func TestValidateSettings_Clean(t *testing.T) {
	issues := ValidateSettings(cleanSettings())
	if HasErrors(issues) {
		t.Fatalf("clean config produced errors: %v", issues)
	}
}

func TestValidateSettings_MissingCoreTemplates(t *testing.T) {
	c := cleanSettings()
	c.DbContextTemplate = "" // unset -> normalized to PathBase in real load; both count as unset
	c.GenerateModels = true
	c.ModelTemplate = "/base" // simulates the empty-path-normalized-to-PathBase case
	c.GenerateProcessors = true
	// ProcessorTemplate left "" (unset)

	issues := ValidateSettings(c)

	if !hasIssue(issues, SeverityError, "DbContextTemplate") {
		t.Error("expected error for missing DbContextTemplate")
	}
	if !hasIssue(issues, SeverityError, "ModelTemplate") {
		t.Error("expected error for missing ModelTemplate (PathBase counts as unset)")
	}
	if !hasIssue(issues, SeverityError, "ProcessorTemplate") {
		t.Error("expected error for missing ProcessorTemplate")
	}
}

func TestValidateSettings_Warnings(t *testing.T) {
	c := cleanSettings()
	c.Mappings = []Mapping{{DatabaseTypes: []string{"text"}, MappedType: "string"}} // no "*"
	c.ClearOutputFolder = true
	c.RemoveOrphanedFiles = true
	c.GenerateProcessorsForVoidReturns = true // but GenerateProcessors is false

	issues := ValidateSettings(c)

	if HasErrors(issues) {
		t.Fatalf("expected only warnings, got errors: %v", issues)
	}
	for _, target := range []string{"Mappings", "RemoveOrphanedFiles", "GenerateProcessorsForVoidReturns"} {
		if !hasIssue(issues, SeverityWarning, target) {
			t.Errorf("expected warning for %s", target)
		}
	}
}

func TestValidateSettings_CopyTargets(t *testing.T) {
	c := cleanSettings()
	c.GenerateCopyTargets = true
	// CopyTargetTemplate unset, no targets
	issues := ValidateSettings(c)
	if !hasIssue(issues, SeverityError, "CopyTargetTemplate") {
		t.Error("expected error for missing CopyTargetTemplate")
	}
	if !hasIssue(issues, SeverityWarning, "CopyTargets") {
		t.Error("expected warning for empty CopyTargets")
	}

	c.CopyTargetTemplate = "/base/copy.gotmpl"
	c.CopyTargets = []CopyTargetConfig{{Table: "t", Format: "bogus"}}
	issues = ValidateSettings(c)
	if !hasIssue(issues, SeverityError, "CopyTargets") {
		t.Error("expected error for invalid copy Format")
	}
}

func TestValidateSettings_AdditionalGenerators(t *testing.T) {
	c := cleanSettings()
	c.AdditionalGenerators = []AdditionalGenerator{
		{Name: "SingleNoFile", Enabled: true, Template: "/base/g.gotmpl", OutputFolder: "/base/x", GenerationType: "single-file"},
		{Name: "BadType", Enabled: true, Template: "/base/g.gotmpl", OutputFolder: "/base/x", GenerationType: "nonsense"},
		{Name: "BadCase", Enabled: true, Template: "/base/g.gotmpl", OutputFolder: "/base/x", GenerationType: "per-routine", FileExtension: ".ts", FileCase: "kebab"},
		{Name: "Disabled", Enabled: false, GenerationType: "single-file"}, // ignored
	}

	issues := ValidateSettings(c)

	if !hasIssue(issues, SeverityError, "AdditionalGenerators") {
		t.Fatal("expected errors from additional generators")
	}
	// The disabled generator must not raise anything about its missing template.
	for _, i := range issues {
		if strings.Contains(i.Message, "Disabled") {
			t.Errorf("disabled generator should be ignored, got: %v", i)
		}
	}
}
