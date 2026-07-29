package dbGen

import (
	"fmt"
	"github.com/guregu/null/v5"
	common2 "github.com/keenmate/db-gen/private/helpers"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var defaultConfigPaths = []string{"./db-gen.json", "./db-gen/db-gen.json", "./db-gen/config.json"}

var localPrefixes = []string{"local.", ".local."}
var localPostfixes = []string{".local"}

type Config struct {
	PathBase                         string                     //for now just using config folder
	ConnectionString                 string                     `mapstructure:"ConnectionString"`
	OutputFolder                     string                     `mapstructure:"OutputFolder"`
	ProcessorsFolderName             string                     `mapstructure:"ProcessorsFolderName"`
	ModelsFolderName                 string                     `mapstructure:"ModelsFolderName"`
	GenerateModels                   bool                       `mapstructure:"GenerateModels"`
	GenerateProcessors               bool                       `mapstructure:"GenerateProcessors"`
	GenerateProcessorsForVoidReturns bool                       `mapstructure:"GenerateProcessorsForVoidReturns"`
	DbContextTemplate                string                     `mapstructure:"DbContextTemplate"`
	ModelTemplate                    string                     `mapstructure:"ModelTemplate"`
	ProcessorTemplate                string                     `mapstructure:"ProcessorTemplate"`
	GeneratedFileExtension           string                     `mapstructure:"GeneratedFileExtension"`
	GeneratedFileCase                string                     `mapstructure:"GeneratedFileCase"`
	Debug                            bool                       `mapstructure:"Debug"`
	ClearOutputFolder                bool                       `mapstructure:"ClearOutputFolder"`
	RemoveOrphanedFiles              bool                       `mapstructure:"RemoveOrphanedFiles"`
	RoutinesFile                     string                     `mapstructure:"RoutinesFile"`
	UseRoutinesFile                  bool                       `mapstructure:"UseRoutinesFile"`
	Generate                         []SchemaConfig             `mapstructure:"Generate"`
	Mappings                         []Mapping                  `mapstructure:"Mappings"`
	UseUserContext                   bool                       `mapstructure:"UseUserContext"`
	UserContextParameterName         string                     `mapstructure:"UserContextParameterName"`
	UserContextType                  string                     `mapstructure:"UserContextType"`
	ContextParameterMappings         []ContextParameterMapping  `mapstructure:"ContextParameterMappings"`
	ParameterSecurityMappings        []ParameterSecurityMapping `mapstructure:"ParameterSecurityMappings"`
	DefaultParameterSecurityLevel    string                     `mapstructure:"DefaultParameterSecurityLevel"`
	AdditionalGenerators             []AdditionalGenerator      `mapstructure:"AdditionalGenerators"`
	Validation                       ValidationConfig           `mapstructure:"Validation"`
	GenerateCopyTargets              bool                       `mapstructure:"GenerateCopyTargets"`
	CopyTargetTemplate               string                     `mapstructure:"CopyTargetTemplate"`
	CopyTargetsFolderName            string                     `mapstructure:"CopyTargetsFolderName"`
	CopyTargets                      []CopyTargetConfig         `mapstructure:"CopyTargets"`
}

// CopyTargetConfig declares a table to generate bulk-COPY code for.
type CopyTargetConfig struct {
	Schema     string `mapstructure:"Schema"`
	Table      string `mapstructure:"Table"`
	MappedName string `mapstructure:"MappedName"` // overrides the generated struct/file name
	Format     string `mapstructure:"Format"`     // "csv" | "text" | "binary" (template hint)
	NullString string `mapstructure:"NullString"` // NULL sentinel for text/csv formats
}

type SchemaConfig struct {
	Schema       string                    `mapstructure:"Schema"`
	AllFunctions bool                      `mapstructure:"AllFunctions"`
	Functions    map[string]RoutineMapping `mapstructure:"Functions"`
}

type RoutineMapping struct {
	Generate            bool
	MappedName          string                   `mapstructure:"MappedName"`
	DontRetrieveValues  bool                     `mapstructure:"DontRetrieveValues"`
	SelectOnlySpecified bool                     `mapstructure:"SelectOnlySpecified"`
	Model               map[string]ColumnMapping `mapstructure:"Model"`
	Parameters          map[string]ParamMapping  `mapstructure:"Parameters"`
}

type ColumnMapping struct {
	SelectColumn    bool
	MappedName      string `mapstructure:"MappedName"`
	MappedType      string `mapstructure:"MappedType"`
	MappingFunction string `mapstructure:"MappingFunction"`

	IsNullable null.Bool `mapstructure:"IsNullable"`
}

type ParamMapping struct {
	MappedName    string    `mapstructure:"MappedName"`
	MappedType    string    `mapstructure:"MappedType"`
	IsNullable    null.Bool `mapstructure:"IsNullable"`
	IsOptional    null.Bool `mapstructure:"IsOptional"`
	SecurityLevel string    `mapstructure:"SecurityLevel"` // none | secure | strict | omit (per-function override)
}

type Mapping struct {
	DatabaseTypes         []string `mapstructure:"DatabaseTypes"`
	MappedType            string   `mapstructure:"MappedType"`
	MappingFunction       string   `mapstructure:"MappingFunction"`
	NullableReturnType    string   `mapstructure:"NullableReturnType"`
	NullableParameterType string   `mapstructure:"NullableParameterType"`
	OptionalParameterType string   `mapstructure:"OptionalParameterType"`
}

type ContextParameterMapping struct {
	ParameterNames []string `mapstructure:"ParameterNames"`
	ContextPath    string   `mapstructure:"ContextPath"`
}

// ParameterSecurityMapping assigns a logging-sensitivity level to parameters by
// name, globally across all routines. Same shape as ContextParameterMapping.
type ParameterSecurityMapping struct {
	ParameterNames []string `mapstructure:"ParameterNames"`
	SecurityLevel  string   `mapstructure:"SecurityLevel"` // none | secure | strict | omit
}

type AdditionalGenerator struct {
	Name              string `mapstructure:"Name"`
	Enabled           bool   `mapstructure:"Enabled"`
	Template          string `mapstructure:"Template"`
	OutputFolder      string `mapstructure:"OutputFolder"`
	FileName          string `mapstructure:"FileName"`
	FileExtension     string `mapstructure:"FileExtension"`
	FileCase          string `mapstructure:"FileCase"`
	GenerationType    string `mapstructure:"GenerationType"`    // "per-routine" or "single-file"
	CleanOutputFolder bool   `mapstructure:"CleanOutputFolder"` // Clean output folder before generation
}

type ValidationBehavior struct {
	CollectAllErrors      bool   `mapstructure:"CollectAllErrors"`
	ReturnType            string `mapstructure:"ReturnType"`
	ErrorResponseTemplate string `mapstructure:"ErrorResponseTemplate"`
}

type ValidationRuleDefinition struct {
	Name              string                 `mapstructure:"Name"`
	Type              string                 `mapstructure:"Type"`         // "Built-in", "Custom", "Generated"
	ExecutorType      string                 `mapstructure:"ExecutorType"` // "ExistingFunction", "GenerateCode"
	ExecutorReference string                 `mapstructure:"ExecutorReference"`
	GeneratedCode     string                 `mapstructure:"GeneratedCode"`
	ErrorMessage      string                 `mapstructure:"ErrorMessage"`
	ErrorCode         string                 `mapstructure:"ErrorCode"`
	Templates         map[string]interface{} `mapstructure:"Templates"` // Strategy -> Template (string or nested map)
	Parameters        map[string]interface{} `mapstructure:"Parameters"`
}

type ParameterValidationMapping struct {
	ParameterNames []string      `mapstructure:"ParameterNames"`
	Rules          []interface{} `mapstructure:"Rules"` // Can be string or map
}

type ValidationConfig struct {
	ValidationStrategy          string                              `mapstructure:"ValidationStrategy"`
	ValidationLocations         []string                            `mapstructure:"ValidationLocations"`
	ValidationBehavior          ValidationBehavior                  `mapstructure:"ValidationBehavior"`
	ValidationRuleDefinitions   []ValidationRuleDefinition          `mapstructure:"ValidationRuleDefinitions"`
	ParameterValidationMappings []ParameterValidationMapping        `mapstructure:"ParameterValidationMappings"`
	FunctionSpecificValidations map[string]map[string][]interface{} `mapstructure:"FunctionSpecificValidations"`
}

// set in ReadConfig
var loadedConfigLocation = ""

// GetAndValidateConfig gets configuration from viper
func GetAndValidateConfig() (*Config, error) {
	config := &Config{
		PathBase:                         "",
		ConnectionString:                 "",
		OutputFolder:                     "",
		ProcessorsFolderName:             "processors",
		ModelsFolderName:                 "models",
		GenerateModels:                   false,
		GenerateProcessors:               false,
		GenerateProcessorsForVoidReturns: false,
		DbContextTemplate:                "",
		ModelTemplate:                    "",
		ProcessorTemplate:                "",
		GeneratedFileExtension:           "",
		GeneratedFileCase:                "",
		Debug:                            false,
		ClearOutputFolder:                false,
		RemoveOrphanedFiles:              false,
		Generate:                         nil,
		Mappings:                         nil,
		RoutinesFile:                     "./db-gen-routines.json",
		UseRoutinesFile:                  false,
		UseUserContext:                   false,
		UserContextParameterName:         "ctx",
		UserContextType:                  "UserContext",
		ContextParameterMappings:         nil,
		ParameterSecurityMappings:        nil,
		DefaultParameterSecurityLevel:    DefaultSecurityLevel,
		AdditionalGenerators:             nil,
		GenerateCopyTargets:              false,
		CopyTargetTemplate:               "",
		CopyTargetsFolderName:            "copy",
		CopyTargets:                      nil,
		Validation: ValidationConfig{
			ValidationStrategy:  "",
			ValidationLocations: nil,
			ValidationBehavior: ValidationBehavior{
				CollectAllErrors:      true,
				ReturnType:            "ValidationResult",
				ErrorResponseTemplate: "",
			},
			ValidationRuleDefinitions:   nil,
			ParameterValidationMappings: nil,
			FunctionSpecificValidations: nil,
		},
	}

	err := getConfigFromViper(config)
	if err != nil {
		return nil, fmt.Errorf("error processing configuration: %s", err)
	}

	// no configuration file loaded
	if loadedConfigLocation == "" {
		return nil, fmt.Errorf("no configuration file loaded")
	}
	// set in TryReadConfigFile
	config.PathBase = filepath.Dir(loadedConfigLocation)

	//All paths are relative to basePath(config file folder)
	config.ProcessorTemplate = joinIfRelative(config.PathBase, config.ProcessorTemplate)
	config.DbContextTemplate = joinIfRelative(config.PathBase, config.DbContextTemplate)
	config.ModelTemplate = joinIfRelative(config.PathBase, config.ModelTemplate)

	config.CopyTargetTemplate = joinIfRelative(config.PathBase, config.CopyTargetTemplate)

	config.OutputFolder = joinIfRelative(config.PathBase, config.OutputFolder)
	// TODO maybe it is better to be relative to Output folder, not Base path
	config.RoutinesFile = joinIfRelative(config.PathBase, config.RoutinesFile)

	// Normalize additional generator paths
	for i := range config.AdditionalGenerators {
		gen := &config.AdditionalGenerators[i]
		gen.Template = joinIfRelative(config.PathBase, gen.Template)
		gen.OutputFolder = joinIfRelative(config.PathBase, gen.OutputFolder)
		gen.FileCase = strings.ToLower(gen.FileCase)

		// Set default generation type if not specified
		if gen.GenerationType == "" {
			if gen.FileName != "" {
				gen.GenerationType = "single-file"
			} else {
				gen.GenerationType = "per-routine"
			}
		}
	}

	config.GeneratedFileCase = strings.ToLower(config.GeneratedFileCase)

	if !common2.Contains(ValidCaseNormalized, config.GeneratedFileCase) {
		return nil, fmt.Errorf(" '%s' is not valid case (maybe GeneratedFileCase is missing)", config.GeneratedFileCase)
	}

	if err := normalizeAndValidateSecurityLevels(config); err != nil {
		return nil, err
	}

	common2.LogDebug("Loaded configuration: \n%+v", config)
	return config, nil
}

func joinIfRelative(basePath string, joiningPath string) string {
	if filepath.IsAbs(joiningPath) {
		return joiningPath
	}

	return filepath.Join(basePath, joiningPath)
}

func ReadConfig(configLocation string) (string, error) {
	// TODO refactor out duplicit code

	// explicitly set configuration
	if configLocation != "" {
		fileExists, err := TryReadConfigFile(configLocation)
		if !fileExists {
			return "", fmt.Errorf("configuration file %s doesnt exist or cannot be read", configLocation)
		}

		if err != nil {
			return "", fmt.Errorf("error reading/parsing configuration file %s: %s", configLocation, err)
		}

		loadedConfigLocation = configLocation

		// load local config

		localConfigExists, err := TryReadLocalConfig(configLocation)
		if err != nil {
			return "", fmt.Errorf("loading local config: %v", err)
		}

		if localConfigExists {
			common2.Log("Local config override loaded")
		}

		return configLocation, nil
	}

	common2.LogDebug("No configuration file set, trying default locations")

	for _, defaultConfigPath := range defaultConfigPaths {
		fileExists, err := TryReadConfigFile(defaultConfigPath)
		if fileExists {
			if err != nil {
				return "", fmt.Errorf("error reading/parsing configuration file %s: %s", configLocation, err)
			}

			loadedConfigLocation = defaultConfigPath

			// load local config
			localConfigExists, err := TryReadLocalConfig(defaultConfigPath)

			if err != nil {
				return "", fmt.Errorf("loading local config: %w", err)
			}

			if localConfigExists {
				common2.Log("Local config override loaded")
			}

			return defaultConfigPath, nil
		}
	}

	// no config file found
	return "", fmt.Errorf("no configuration file set and no file found at default locations (see readme)")
}

func TryReadLocalConfig(configLocation string) (bool, error) {
	common2.LogDebug("Checking if local config exists")

	for _, path := range getPossibleLocalConfigs(configLocation) {
		exists, err := TryReadConfigFile(path)

		if exists {
			common2.LogDebug("Local config at %s loaded", path)
			return exists, err
		}
	}

	return false, nil
}

func TryReadConfigFile(configPath string) (bool, error) {
	common2.LogDebug("Trying to read config file: %s", configPath)

	// TODO this could hide some usefull errors, maybe log the reason in debug mode
	if !common2.FileIsReadable(configPath) {
		return false, nil
	}

	file, err := os.Open(configPath)
	defer file.Close()
	if err != nil {
		return true, fmt.Errorf("opening file: %s", err)
	}

	viper.SetConfigType(filepath.Ext(configPath)[1:])

	err = viper.MergeConfig(file)
	if err != nil {
		return true, fmt.Errorf("reading configuration: %s", err)
	}
	common2.LogDebug("Configuration file at %s loaded", configPath)

	return true, nil
}

func getPossibleLocalConfigs(configLocation string) []string {
	paths := make([]string, 0)

	directory := filepath.Dir(configLocation)
	file := filepath.Base(configLocation)
	fileWithoutExtension := strings.TrimSuffix(file, filepath.Ext(configLocation))

	// prefixes
	for _, prefix := range localPrefixes {
		paths = append(paths, filepath.Join(directory, prefix+file))
	}

	// postfixes
	for _, postfix := range localPostfixes {
		paths = append(paths, filepath.Join(directory, fileWithoutExtension+postfix))
		paths = append(paths, filepath.Join(directory, file+postfix))
	}

	return paths
}
