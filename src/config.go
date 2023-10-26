package dbGen

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const defaultConfigPath = "./db-gen.json"

type cliArgs struct {
	command    Command
	configPath string
	verbose    bool
}

type Config struct {
	PathBase                   string         //for now just using config folder
	ConnectionString           string         `json:"ConnectionString"`
	OutputFolder               string         `json:"OutputFolder,omitempty"`
	GenerateModels             bool           `json:"GenerateModels,omitempty"`
	GenerateProcessors         bool           `json:"GenerateProcessors,omitempty"`
	SkipModelGenForVoidReturns bool           `json:"SkipModelGenForVoidReturns,omitempty"`
	DbContextTemplate          string         `json:"DbContextTemplate,omitempty"`
	ModelTemplate              string         `json:"ModelTemplate,omitempty"`
	ProcessorTemplate          string         `json:"ProcessorTemplate,omitempty"`
	GeneratedFileExtension     string         `json:"GeneratedFileExtension,omitempty"`
	GeneratedFileCase          string         `json:"GeneratedFileCase,omitempty"`
	Verbose                    bool           `json:"Verbose,omitempty"`
	ClearOutputFolder          bool           `json:"ClearOutputFolder,omitempty"`
	Generate                   []SchemaConfig `json:"Generate,omitempty"`
	Mappings                   []Mapping      `json:"Mappings"`
}

type SchemaConfig struct {
	Schema           string   `json:"Schema,omitempty"`
	AllFunctions     bool     `json:"AllFunctions,omitempty"`
	Functions        []string `json:"Functions,omitempty"`
	IgnoredFunctions []string `json:"IgnoredFunctions,omitempty"`
}

type Mapping struct {
	DatabaseTypes   []string `json:"DatabaseTypes"`
	MappedType      string   `json:"MappedType"`
	MappingFunction string   `json:"MappingFunction"`
}

type Command string

const (
	Init    Command = "init"
	Gen             = "gen"
	Version         = "version"
)

var CurrentConfig *Config = nil

func GetCommand() Command {
	// ignore error because we
	args := parseCLIArgs()
	log.Printf("executing command %s \n", args.command)
	return args.command
}

func GetConfig() (*Config, error) {
	args := parseCLIArgs()

	config, err := readJsonConfigFile(args.configPath)

	if err != nil {
		return nil, fmt.Errorf("getting configuration from file: %w", err)
	}

	// TODO Allow some config values (connection_string) from separate file

	// Cli args should override config loaded from file
	config.Verbose = args.verbose
	config.PathBase = filepath.Dir(args.configPath)
	//All paths are relative to config file
	config.ProcessorTemplate = joinIfRelative(config.PathBase, config.ProcessorTemplate)
	config.DbContextTemplate = joinIfRelative(config.PathBase, config.DbContextTemplate)
	config.ModelTemplate = joinIfRelative(config.PathBase, config.ModelTemplate)
	config.OutputFolder = joinIfRelative(config.PathBase, config.OutputFolder)

	if !contains(ValidCase, config.GeneratedFileCase) {
		return nil, fmt.Errorf("%s is not valid case", config.GeneratedFileCase)
	}

	CurrentConfig = config

	VerboseLog("%+v", config)
	return config, nil
}

func readJsonConfigFile(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	config := new(Config)

	err = json.Unmarshal(file, config)

	if err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	return config, nil
}

var args *cliArgs

// TODO refactor to separate parsing and validating cli
func parseCLIArgs() *cliArgs {
	// only parse args once
	if args != nil {
		return args
	}

	verboseFlag := flag.Bool("verbose", false, "If true it will print more stuff")
	configPathFlag := flag.String("config", defaultConfigPath, "Path to config file, all paths are relative it")
	flag.Parse()

	args = new(cliArgs)
	args.command = parseCommand(flag.Arg(0))
	args.verbose = *verboseFlag
	args.configPath = *configPathFlag
	return args
}

func parseCommand(command string) Command {
	switch strings.ToLower(command) {
	case "gen":
		return Gen
	case "init":
		return Init
	case "version":
		return Version
	default:
		return Gen
	}
}

func joinIfRelative(basePath string, joiningPath string) string {
	if filepath.IsAbs(joiningPath) {
		return joiningPath
	}

	return filepath.Join(basePath, joiningPath)
}
