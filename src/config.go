package dbGen

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

	// Cli args should override config loaded from file
	config.Command = args.command
	config.Verbose = args.verbose

	if err != nil {
		return nil, fmt.Errorf("getting configuration from file: %w", err)
	}

	CurrentConfig = config
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
	configPathFlag := flag.String("config", defaultConfigPath, "If true it will print more stuff")
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
