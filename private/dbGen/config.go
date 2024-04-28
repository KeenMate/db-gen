package dbGen

import (
	"fmt"
	"github.com/guregu/null/v5"
	common2 "github.com/keenmate/db-gen/private/common"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var defaultConfigPaths = []string{"./db-gen.json", "./db-gen/db-gen.json", "./db-gen/config.json"}

var localPrefixes = []string{"local.", ".local."}
var localPostfixes = []string{".local"}

type Config struct {
	PathBase                         string         //for now just using config folder
	ConnectionString                 string         `mapstructure:"ConnectionString"`
	OutputFolder                     string         `mapstructure:"OutputFolder"`
	ProcessorsFolderName             string         `mapstructure:"ProcessorsFolderName"`
	ModelsFolderName                 string         `mapstructure:"ModelsFolderName"`
	GenerateModels                   bool           `mapstructure:"GenerateModels"`
	GenerateProcessors               bool           `mapstructure:"GenerateProcessors"`
	GenerateProcessorsForVoidReturns bool           `mapstructure:"GenerateProcessorsForVoidReturns"`
	DbContextTemplate                string         `mapstructure:"DbContextTemplate"`
	ModelTemplate                    string         `mapstructure:"ModelTemplate"`
	ProcessorTemplate                string         `mapstructure:"ProcessorTemplate"`
	GeneratedFileExtension           string         `mapstructure:"GeneratedFileExtension"`
	GeneratedFileCase                string         `mapstructure:"GeneratedFileCase"`
	Debug                            bool           `mapstructure:"Debug"`
	ClearOutputFolder                bool           `mapstructure:"ClearOutputFolder"`
	RoutinesFile                     string         `mapstructure:"RoutinesFile"`
	UseRoutinesFile                  bool           `mapstructure:"UseRoutinesFile"`
	Generate                         []SchemaConfig `mapstructure:"Generate"`
	Mappings                         []Mapping      `mapstructure:"Mappings"`
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
	MappedName string    `mapstructure:"MappedName"`
	MappedType string    `mapstructure:"MappedType"`
	IsNullable null.Bool `mapstructure:"IsNullable"`
	IsOptional null.Bool `mapstructure:"IsOptional"`
}

type Mapping struct {
	DatabaseTypes   []string `mapstructure:"DatabaseTypes"`
	MappedType      string   `mapstructure:"MappedType"`
	MappingFunction string   `mapstructure:"MappingFunction"`
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
		Generate:                         nil,
		Mappings:                         nil,
		RoutinesFile:                     "./db-gen-routines.json",
		UseRoutinesFile:                  false,
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
	config.OutputFolder = joinIfRelative(config.PathBase, config.OutputFolder)
	config.RoutinesFile = joinIfRelative(config.PathBase, config.RoutinesFile)

	config.GeneratedFileCase = strings.ToLower(config.GeneratedFileCase)

	if !common2.Contains(ValidCaseNormalized, config.GeneratedFileCase) {
		return nil, fmt.Errorf(" '%s' is not valid case (maybe GeneratedFileCase is missing)", config.GeneratedFileCase)
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
