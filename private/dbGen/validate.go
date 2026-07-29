package dbGen

import (
	"fmt"
	common2 "github.com/keenmate/db-gen/private/helpers"
	"github.com/keenmate/db-gen/private/version"
	"io"
	"path/filepath"
	"strings"
	"text/template"
)

// Severity of a validation issue.
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "error"
	SeverityWarning IssueSeverity = "warning"
)

// ValidationIssue is a single problem found while validating settings or templates.
type ValidationIssue struct {
	Severity IssueSeverity
	Area     string // "settings" or "template"
	Target   string // the config key / template involved (optional)
	Message  string
}

func (i ValidationIssue) String() string {
	target := ""
	if i.Target != "" {
		target = " [" + i.Target + "]"
	}
	return fmt.Sprintf("%-7s %s%s: %s", strings.ToUpper(string(i.Severity)), i.Area, target, i.Message)
}

// pathUnset reports whether a path-typed config value was left empty. Empty path
// values are normalized to the config directory by joinIfRelative during load,
// so an unset template resolves to PathBase rather than "".
func pathUnset(config *Config, p string) bool {
	return p == "" || filepath.Clean(p) == filepath.Clean(config.PathBase)
}

// HasErrors reports whether any issue in the slice is error-severity.
func HasErrors(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// ValidateSettings performs static checks on an already-loaded config. It does
// not connect to the database and does not parse templates. It collects every
// problem it finds rather than stopping at the first.
//
// GetAndValidateConfig has already enforced the hard requirements (config loads,
// GeneratedFileCase and security levels are valid); these checks catch the
// remaining "you enabled X but didn't configure Y" style mistakes.
func ValidateSettings(config *Config) []ValidationIssue {
	var issues []ValidationIssue
	add := func(sev IssueSeverity, target, format string, args ...any) {
		issues = append(issues, ValidationIssue{sev, "settings", target, fmt.Sprintf(format, args...)})
	}

	if config.GeneratedFileExtension == "" {
		add(SeverityWarning, "GeneratedFileExtension", "no file extension set; generated files will have no extension")
	}

	if pathUnset(config, config.DbContextTemplate) {
		add(SeverityError, "DbContextTemplate", "no DbContext template set (always generated)")
	}
	if config.GenerateModels && pathUnset(config, config.ModelTemplate) {
		add(SeverityError, "ModelTemplate", "GenerateModels is true but no model template is set")
	}
	if config.GenerateProcessors && pathUnset(config, config.ProcessorTemplate) {
		add(SeverityError, "ProcessorTemplate", "GenerateProcessors is true but no processor template is set")
	}
	if config.GenerateProcessorsForVoidReturns && !config.GenerateProcessors {
		add(SeverityWarning, "GenerateProcessorsForVoidReturns", "has no effect while GenerateProcessors is false")
	}

	if len(config.Generate) == 0 {
		add(SeverityError, "Generate", "no schemas configured; nothing to generate")
	}
	if len(config.Mappings) == 0 {
		add(SeverityError, "Mappings", "no type mappings configured")
	} else {
		hasCatchAll := false
		for _, m := range config.Mappings {
			if common2.Contains(m.DatabaseTypes, "*") {
				hasCatchAll = true
				break
			}
		}
		if !hasCatchAll {
			add(SeverityWarning, "Mappings", `no "*" catch-all mapping; an unmapped database type will stop generation`)
		}
	}

	if config.ClearOutputFolder && config.RemoveOrphanedFiles {
		add(SeverityWarning, "RemoveOrphanedFiles", "ignored while ClearOutputFolder is true (orphan removal only runs when ClearOutputFolder is false)")
	}

	// Copy targets
	if config.GenerateCopyTargets {
		if pathUnset(config, config.CopyTargetTemplate) {
			add(SeverityError, "CopyTargetTemplate", "GenerateCopyTargets is true but no copy-target template is set")
		}
		if len(config.CopyTargets) == 0 {
			add(SeverityWarning, "CopyTargets", "GenerateCopyTargets is true but no target tables are listed")
		}
		for _, t := range config.CopyTargets {
			if t.Table == "" {
				add(SeverityError, "CopyTargets", "a copy target has no Table set")
			}
			if t.Format != "" && !common2.Contains([]string{"csv", "text", "binary"}, strings.ToLower(t.Format)) {
				add(SeverityError, "CopyTargets", "copy target %q has invalid Format %q (expected csv/text/binary)", t.Table, t.Format)
			}
		}
	}

	// Additional generators
	for _, gen := range config.AdditionalGenerators {
		if !gen.Enabled {
			continue
		}
		label := gen.Name
		if label == "" {
			label = "(unnamed)"
		}
		if pathUnset(config, gen.Template) {
			add(SeverityError, "AdditionalGenerators", "generator %q has no Template set", label)
		}
		switch gen.GenerationType {
		case "single-file":
			if gen.FileName == "" {
				add(SeverityError, "AdditionalGenerators", "single-file generator %q has no FileName set", label)
			}
		case "per-routine":
			if gen.FileExtension == "" {
				add(SeverityWarning, "AdditionalGenerators", "per-routine generator %q has no FileExtension set", label)
			}
			if gen.FileCase == "" {
				add(SeverityWarning, "AdditionalGenerators", "per-routine generator %q has no FileCase set; file names fall back to the raw routine name", label)
			}
		default:
			add(SeverityError, "AdditionalGenerators", "generator %q has invalid GenerationType %q (expected single-file/per-routine)", label, gen.GenerationType)
		}
		if gen.FileCase != "" && !common2.Contains(ValidCaseNormalized, gen.FileCase) {
			add(SeverityError, "AdditionalGenerators", "generator %q has invalid FileCase %q", label, gen.FileCase)
		}
	}

	return issues
}

// ValidateTemplates parses every template that would actually be used, and — when
// canExecute is true — renders each against the supplied data to a discard writer
// to surface field/method errors that only appear at execution time. Pass the
// processed routines and copy targets from the normal pipeline; when canExecute
// is false (no database and no routines file), only parsing is checked.
func ValidateTemplates(config *Config, routines []Routine, copyTargets []CopyTarget, canExecute bool) []ValidationIssue {
	var issues []ValidationIssue
	add := func(sev IssueSeverity, target, format string, args ...any) {
		issues = append(issues, ValidationIssue{sev, "template", target, fmt.Sprintf(format, args...)})
	}

	buildInfo := version.GetBuildInfo()

	// parseOnce parses a template and returns it, recording a parse error. Unset
	// paths are skipped silently — ValidateSettings already flags required ones.
	parseOnce := func(label, path string) *template.Template {
		if pathUnset(config, path) {
			return nil
		}
		tmpl, err := parseTemplate(path)
		if err != nil {
			add(SeverityError, label, "parse failed: %s", err)
			return nil
		}
		return tmpl
	}

	// execCheck renders a template to io.Discard, recording an execution error.
	execCheck := func(label string, tmpl *template.Template, data interface{}) {
		if tmpl == nil || !canExecute {
			return
		}
		if err := tmpl.Execute(io.Discard, data); err != nil {
			add(SeverityError, label, "render failed: %s", err)
		}
	}

	// DbContext (always generated)
	if dbCtx := parseOnce("DbContextTemplate", config.DbContextTemplate); dbCtx != nil {
		execCheck("DbContextTemplate", dbCtx, &DbContextData{Config: config, Functions: routines, BuildInfo: buildInfo})
	}

	if config.GenerateModels {
		if modelTmpl := parseOnce("ModelTemplate", config.ModelTemplate); modelTmpl != nil {
			for _, r := range routines {
				if !r.HasReturn {
					continue
				}
				execCheck("ModelTemplate", modelTmpl, &ModelTemplateData{Config: config, Routine: r, BuildInfo: buildInfo})
			}
		}
	}

	if config.GenerateProcessors {
		if procTmpl := parseOnce("ProcessorTemplate", config.ProcessorTemplate); procTmpl != nil {
			for _, r := range routines {
				if !config.GenerateProcessorsForVoidReturns && !r.HasReturn {
					continue
				}
				execCheck("ProcessorTemplate", procTmpl, &ProcessorTemplateData{Config: config, Routine: r, BuildInfo: buildInfo})
			}
		}
	}

	if config.GenerateCopyTargets {
		if copyTmpl := parseOnce("CopyTargetTemplate", config.CopyTargetTemplate); copyTmpl != nil {
			for _, t := range copyTargets {
				execCheck("CopyTargetTemplate", copyTmpl, &CopyTargetTemplateData{Config: config, CopyTarget: t, BuildInfo: buildInfo})
			}
		}
	}

	for _, gen := range config.AdditionalGenerators {
		if !gen.Enabled || pathUnset(config, gen.Template) {
			continue
		}
		label := "AdditionalGenerator:" + gen.Name
		tmpl := parseOnce(label, gen.Template)
		if tmpl == nil {
			continue
		}
		if gen.GenerationType == "single-file" {
			execCheck(label, tmpl, &DbContextData{Config: config, Functions: routines, BuildInfo: buildInfo})
		} else {
			for _, r := range routines {
				// mirror generatePerRoutineFiles' TypeScript skip
				if !r.HasReturn && gen.Name == "TypeScript" {
					continue
				}
				execCheck(label, tmpl, &ModelTemplateData{Config: config, Routine: r, BuildInfo: buildInfo})
			}
		}
	}

	return issues
}
