package cmd

import (
	"github.com/keenmate/db-gen/private/helpers"
)

const (
	keyDebug            = "debug"
	keyConnectionString = "connectionString"
	keyConfig           = "config"
)

var commonFlags = []helpers.FlagArgument{
	helpers.NewBoolFlag(keyDebug, "d", false, "Print debug logs and create debug files"),
	helpers.NewStringFlag(keyConfig, "s", "", "Connection string used to connect to database"),
	helpers.NewStringFlag(keyConnectionString, "c", "", "Path to configuration file"),
}

func printDatabaseChanges(databaseChanges string) {
	if len(databaseChanges) == 0 {
		helpers.LogBold("No database changes detected")
		return
	}

	helpers.LogBold("Database changes:\n" + databaseChanges)
}
