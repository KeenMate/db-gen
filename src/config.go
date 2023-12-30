package dbGen

import (
	"fmt"
	"github.com/keenmate/db-gen/common"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var defaultConfigPaths = []string{"./db-gen.json", "./db-gen/db-gen.json", "./db-gen/config.json"}

type Config struct {
	PathBase                         string         //for now just using config folder
	ConnectionString                 string         `mapstructure:"ConnectionString"`
	OutputFolder                     string         `mapstructure:"OutputFolder"`
	GenerateModels                   bool           `mapstructure:"GenerateModels"`
	GenerateProcessors               bool           `mapstructure:"GenerateProcessors"`
	GenerateProcessorsForVoidReturns bool           `mapstructure:"GenerateProcessorsForVoidReturns"`
	DbContextTemplate                string         `mapstructure:"DbContextTemplate"`
	ModelTemplate                    string         `mapstructure:"ModelTemplate"`
	ProcessorTemplate                string         `mapstructure:"ProcessorTemplate"`
	GeneratedFileExtension           string         `mapstructure:"GeneratedFileExtension"`
	GeneratedFileCase                string         `mapstructure:"GeneratedFileCase"`
	Verbose                          bool           `mapstructure:"Verbose"`
	ClearOutputFolder                bool           `mapstructure:"ClearOutputFolder"`
	Generate                         []SchemaConfig `mapstructure:"Generate"`
	Mappings                         []Mapping      `mapstructure:"Mappings"`
}

type SchemaConfig struct {
	Schema           string   `mapstructure:"Schema"`
	AllFunctions     bool     `mapstructure:"AllFunctions"`
	Functions        []string `mapstructure:"Functions"`
	IgnoredFunctions []string `mapstructure:"IgnoredFunctions"`
}

type Mapping struct {
	DatabaseTypes   []string `mapstructure:"DatabaseTypes"`
	MappedType      string   `mapstructure:"MappedType"`
	MappingFunction string   `mapstructure:"MappingFunction"`
}

var CurrentConfig *Config = nil

// set in TryReadConfigFile
var loadedConfigLocation string = ""

// GetConfig gets configuration from viper
func GetConfig() (*Config, error) {
	config := new(Config)

	err := viper.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("error mapping configuration: %s", err)
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

	if !contains(ValidCase, strings.ToLower(config.GeneratedFileCase)) {
		return nil, fmt.Errorf(" '%s' is not valid case (maybe GeneratedFileCase is missing)", config.GeneratedFileCase)
	}

	// used by debug helpers
	CurrentConfig = config

	common.LogDebug("%+v", config)
	return config, nil
}

func joinIfRelative(basePath string, joiningPath string) string {
	if filepath.IsAbs(joiningPath) {
		return joiningPath
	}

	return filepath.Join(basePath, joiningPath)
}

func ReadConfig(configLocation string) (string, error) {
	// explicitly set configuration
	if configLocation != "" {
		err, fileExists := TryReadConfigFile(configLocation)
		if !fileExists {
			return "", fmt.Errorf("configuration file %s doesnt exist or cannot be read", configLocation)
		}

		if err != nil {
			return "", fmt.Errorf("error reading/parsing configuration file %s: %s", configLocation, err)
		}

		return configLocation, nil
	}

	for _, defaultConfigPath := range defaultConfigPaths {
		err, exists := TryReadConfigFile(defaultConfigPath)
		if exists {
			if err != nil {
				return "", fmt.Errorf("error reading/parsing configuration file %s: %s", configLocation, err)
			}

			return defaultConfigPath, nil
		}
	}

	// no config file found
	return "", fmt.Errorf("no configuration file set and no file found at default locations (see readme)")
}

func TryReadConfigFile(configPath string) (error, bool) {
	common.LogDebug("Trying to read config file: %s", configPath)

	if !common.PathExists(configPath) {

		return nil, false
	}

	file, err := os.Open(configPath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("opening file: %s", err), true
	}

	viper.SetConfigType(filepath.Ext(configPath)[1:])

	err = viper.MergeConfig(file)
	if err != nil {
		return fmt.Errorf("reading configuration: %s", err), true
	}
	common.LogDebug("%s loaded", configPath)
	// so we can use basePath
	loadedConfigLocation = configPath
	return nil, true
}
