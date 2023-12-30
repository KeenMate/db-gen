package common

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ConfigurationString(command *cobra.Command, key string, shorthand string, defaultValue string, usage string) {
	command.Flags().StringP(key, shorthand, defaultValue, usage)
	_ = viper.BindPFlag(key, command.Flags().Lookup(key))
}
func ConfigurationBool(command *cobra.Command, key string, shorthand string, defaultValue bool, usage string) {
	command.Flags().BoolP(key, shorthand, defaultValue, usage)
	_ = viper.BindPFlag(key, command.Flags().Lookup(key))
}
