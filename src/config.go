package dbGen

import (
	"encoding/json"
	"flag"
	"fmt"
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

var CurrentConfig *Config = nil

func GetConfig() (*Config, error) {
	args, err := parseCLIArgs()
	if err != nil {
		return nil, fmt.Errorf("parsing cli args: %w", err)
	}
	config, err := readJsonConfigFile(args.configPath)

	if err != nil {
		return nil, fmt.Errorf("getting configuration from file: %w", err)
	}
	// TODO Allow some config values (connection_string) from separate file

	// Cli args should override config loaded from file
	config.Command = args.command
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

func parseCLIArgs() (*cliArgs, error) {

	verboseFlag := flag.Bool("verbose", false, "If true it will print more stuff")
	configPathFlag := flag.String("config", defaultConfigPath, "Path to config file, all paths are relative it")
	flag.Parse()

	args := new(cliArgs)
	args.verbose = *verboseFlag
	args.configPath = *configPathFlag

	if _, err := os.Stat(args.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", args.configPath)
	}

	args.command = parseCommand(flag.Arg(0))

	return args, nil
}

func parseCommand(command string) Command {
	switch strings.ToLower(command) {
	case "gen":
		return Gen
	case "init":
		return Init
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
