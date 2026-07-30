package dbGen

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestJoinIfRelative(t *testing.T) {
	abs := t.TempDir() // genuinely absolute on the current OS
	if got := joinIfRelative("/base", abs); got != abs {
		t.Errorf("absolute path should be returned unchanged: got %q", got)
	}
	got := joinIfRelative("base", "templates/x.gotmpl")
	want := filepath.Join("base", "templates/x.gotmpl")
	if got != want {
		t.Errorf("joinIfRelative = %q, want %q", got, want)
	}
}

func TestGetPossibleLocalConfigs(t *testing.T) {
	paths := getPossibleLocalConfigs(filepath.Join("test", "db-gen.json"))
	found := false
	want := filepath.Join("test", "local.db-gen.json")
	for _, p := range paths {
		if p == want {
			found = true
		}
	}
	if !found {
		t.Errorf("expected %q among local config candidates %v", want, paths)
	}
}

// TestTemplateVariablesDecode verifies the passthrough map round-trips through
// viper decoding when present, and defaults to nil when absent. viper is a global
// singleton, so we reset it around each case.
func TestTemplateVariablesDecode(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		viper.Set("TemplateVariables", map[string]interface{}{
			"GeneratedNs":   "Keenmate.MyApp.Database.Generated",
			"UserContextNs": "Keenmate.MyApp.Web.Models",
		})

		config := &Config{}
		if err := getConfigFromViper(config); err != nil {
			t.Fatalf("getConfigFromViper: %v", err)
		}
		if len(config.TemplateVariables) != 2 {
			t.Fatalf("TemplateVariables = %+v, want 2 entries", config.TemplateVariables)
		}
		// viper lowercases config keys on load; values keep their case. The
		// templateVar helper hides this from templates.
		if got := config.TemplateVariables["generatedns"]; got != "Keenmate.MyApp.Database.Generated" {
			t.Errorf("generatedns = %q, want Keenmate.MyApp.Database.Generated", got)
		}
	})

	t.Run("absent", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		config := &Config{}
		if err := getConfigFromViper(config); err != nil {
			t.Fatalf("getConfigFromViper: %v", err)
		}
		if config.TemplateVariables != nil {
			t.Errorf("absent TemplateVariables should be nil, got %+v", config.TemplateVariables)
		}
	})
}

// TestLoadConfig exercises the real config pipeline (ReadConfig + GetAndValidateConfig)
// against the committed copy-target test config. viper is a global singleton, so we
// reset it first to avoid pollution from other tests.
func TestLoadConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	configPath := filepath.Join("..", "..", "test", "db-gen-copy.json")
	if _, err := ReadConfig(configPath); err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	config, err := GetAndValidateConfig()
	if err != nil {
		t.Fatalf("GetAndValidateConfig: %v", err)
	}

	if !config.GenerateCopyTargets {
		t.Error("GenerateCopyTargets should be true")
	}
	if len(config.CopyTargets) != 1 || config.CopyTargets[0].Table != "data_to_process" {
		t.Fatalf("CopyTargets = %+v", config.CopyTargets)
	}
	if config.GeneratedFileCase != "pascalcase" {
		t.Errorf("GeneratedFileCase = %q, want pascalcase", config.GeneratedFileCase)
	}
	// paths are normalized relative to the config's directory
	if filepath.Base(config.CopyTargetTemplate) != "copy-pgx.gotmpl" {
		t.Errorf("CopyTargetTemplate not normalized: %q", config.CopyTargetTemplate)
	}
	if filepath.Base(config.OutputFolder) != "cli-output" {
		t.Errorf("OutputFolder not normalized: %q", config.OutputFolder)
	}
}
