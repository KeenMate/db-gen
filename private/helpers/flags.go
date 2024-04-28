package helpers

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagArgument interface {
	DefineFlag(command *cobra.Command)
	BindFlag(command *cobra.Command)
}

type StringFlag struct {
	key          string
	shorthand    string
	defaultValue string
	usage        string
}

func (f *StringFlag) DefineFlag(command *cobra.Command) {
	command.Flags().StringP(f.key, f.shorthand, f.defaultValue, f.usage)

}

func (f *StringFlag) BindFlag(command *cobra.Command) {
	_ = viper.BindPFlag(f.key, command.Flags().Lookup(f.key))
}

func NewStringFlag(key string, shorthand string, defaultValue string, usage string) *StringFlag {
	return &StringFlag{
		key:          key,
		shorthand:    shorthand,
		defaultValue: defaultValue,
		usage:        usage,
	}
}

type BoolFlag struct {
	key          string
	shorthand    string
	defaultValue bool
	usage        string
}

func (f *BoolFlag) DefineFlag(command *cobra.Command) {
	command.Flags().BoolP(f.key, f.shorthand, f.defaultValue, f.usage)
}

func (f *BoolFlag) BindFlag(command *cobra.Command) {
	_ = viper.BindPFlag(f.key, command.Flags().Lookup(f.key))
}

func NewBoolFlag(key string, shorthand string, defaultValue bool, usage string) *BoolFlag {
	return &BoolFlag{
		key:          key,
		shorthand:    shorthand,
		defaultValue: defaultValue,
		usage:        usage,
	}
}

// BindFlags we nned to separate binding from declaration if we dont have unique name for each flag
func BindFlags(command *cobra.Command, flags []FlagArgument) {
	for _, flag := range flags {
		flag.BindFlag(command)
	}
}

func DefineFlags(command *cobra.Command, flags []FlagArgument) {
	for _, flag := range flags {
		flag.DefineFlag(command)
	}
}
