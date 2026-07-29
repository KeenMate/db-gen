package cmd

import (
	"fmt"
	"github.com/keenmate/db-gen/private/dbGen"
	"github.com/keenmate/db-gen/private/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

const keyOffline = "offline"

var validateFlags = []helpers.FlagArgument{
	helpers.NewBoolFlag(keyUseRoutinesFile, "", false, "Use routines file instead of a database connection"),
	helpers.NewBoolFlag(keyOffline, "", false, "Only validate settings and parse templates; never connect to a database or render templates"),
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and templates",
	Long: `
Validate the configuration file and the templates it references without
generating any files.

Checks, in order:
  1. Settings   - the config loads and enabled features are configured
                  consistently (required templates set, valid enums, etc.).
  2. Templates  - every template that would be used parses (syntax and
                  known functions).
  3. Rendering  - if a database is reachable (or UseRoutinesFile is set, and
                  unless --offline), each template is rendered against your
                  real routines to a discard writer to catch field/method
                  errors that only surface at render time.

Exits non-zero if any error-level problem is found.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		helpers.BindFlags(cmd, append(commonFlags, validateFlags...))
		_, err := dbGen.ReadConfig(viper.GetString(keyConfig))
		if err != nil {
			helpers.Exit("configuration error: %s", err)
		}

		viper.AutomaticEnv() // read in environment variables that match

		if err = doValidate(); err != nil {
			helpers.Exit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	helpers.DefineFlags(validateCmd, append(commonFlags, validateFlags...))
}

func doValidate() error {
	config, err := dbGen.GetAndValidateConfig()
	if err != nil {
		// GetAndValidateConfig failing is itself a settings error; report it
		// as a validation failure rather than a generic crash.
		helpers.LogError("SETTINGS: %s", err)
		return fmt.Errorf("configuration is invalid")
	}

	var issues []dbGen.ValidationIssue

	// 1. Settings
	issues = append(issues, dbGen.ValidateSettings(config)...)

	// 2 + 3. Templates (parse always; render if we can get routines)
	routines, copyTargets, canExecute := loadDataForValidation(config)
	issues = append(issues, dbGen.ValidateTemplates(config, routines, copyTargets, canExecute)...)

	return reportValidation(issues, canExecute)
}

// loadDataForValidation tries to obtain processed routines and copy targets so
// templates can be render-checked. On any failure (e.g. no database) it returns
// canExecute=false so validation degrades to parse-only instead of failing.
func loadDataForValidation(config *dbGen.Config) ([]dbGen.Routine, []dbGen.CopyTarget, bool) {
	if viper.GetBool(keyOffline) {
		helpers.Log("Offline mode: skipping database connection and template rendering")
		return nil, nil, false
	}

	dbRoutines, err := dbGen.GetRoutines(config)
	if err != nil {
		helpers.LogWarn("Could not load routines (%s); templates will be parsed but not rendered", err)
		return nil, nil, false
	}

	if err = dbGen.PreprocessRoutines(&dbRoutines, config); err != nil {
		helpers.LogWarn("Could not preprocess routines (%s); templates will be parsed but not rendered", err)
		return nil, nil, false
	}

	routines, err := dbGen.Process(dbRoutines, config)
	if err != nil {
		helpers.LogWarn("Could not process routines (%s); templates will be parsed but not rendered", err)
		return nil, nil, false
	}

	var copyTargets []dbGen.CopyTarget
	if config.GenerateCopyTargets {
		copyTables, err := dbGen.GetCopyTargets(config)
		if err != nil {
			helpers.LogWarn("Could not load copy targets (%s); copy template will be parsed but not rendered", err)
		} else if copyTargets, err = dbGen.MapCopyTargets(copyTables, config); err != nil {
			helpers.LogWarn("Could not map copy targets (%s); copy template will be parsed but not rendered", err)
		}
	}

	return routines, copyTargets, true
}

func reportValidation(issues []dbGen.ValidationIssue, rendered bool) error {
	errorCount, warnCount := 0, 0
	for _, issue := range issues {
		if issue.Severity == dbGen.SeverityError {
			errorCount++
			helpers.LogError("%s", issue.String())
		} else {
			warnCount++
			helpers.LogWarn("%s", issue.String())
		}
	}

	scope := "settings, templates parsed and rendered"
	if !rendered {
		scope = "settings and templates parsed (rendering skipped)"
	}
	log.Printf("Validation complete (%s): %d error(s), %d warning(s)", scope, errorCount, warnCount)

	if errorCount > 0 {
		return fmt.Errorf("validation failed with %d error(s)", errorCount)
	}
	helpers.LogBold("Configuration and templates are valid")
	return nil
}
