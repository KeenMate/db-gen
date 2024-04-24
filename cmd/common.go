package cmd

import "github.com/keenmate/db-gen/common"

const (
	keyDebug            = "debug"
	keyConnectionString = "connectionString"
	keyConfig           = "config"
)

var commonFlags = []common.FlagArgument{
	common.NewBoolFlag(keyDebug, "d", false, "Print debug logs and create debug files"),
	common.NewStringFlag(keyConfig, "s", "", "Connection string used to connect to database"),
	common.NewStringFlag(keyConnectionString, "c", "", "Path to configuration file"),
}
