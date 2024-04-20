package common

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func DefineStringFlag(command *cobra.Command, key string, shorthand string, defaultValue string, usage string) {
	command.Flags().StringP(key, shorthand, defaultValue, usage)
}

func DefineBoolFlag(command *cobra.Command, key string, shorthand string, defaultValue bool, usage string) {
	command.Flags().BoolP(key, shorthand, defaultValue, usage)
}

// Due to bug/design flaw in viper, we need to bind flags only after we run function

func BindStringFlag(command *cobra.Command, key string) {
	_ = viper.BindPFlag(key, command.Flags().Lookup(key))
}

func BindBoolFlag(command *cobra.Command, key string) {
	_ = viper.BindPFlag(key, command.Flags().Lookup(key))
}
