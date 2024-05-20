package dbGen

import (
	"fmt"
	"github.com/guregu/null/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
)

func getConfigFromViper(configWithDefaultValue *Config) error {
	configOption := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapCustomTypes()))

	err := viper.Unmarshal(configWithDefaultValue, configOption)
	if err != nil {
		return fmt.Errorf("getting values from viper: %s", err)
	}

	return nil
}

func decodeWithHook(in interface{}, out interface{}) error {
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapCustomTypes())
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{DecodeHook: decodeHook, Result: out})
	if err != nil {
		return err
	}

	err = decoder.Decode(in)
	if err != nil {
		return err
	}

	return nil

}

// mapCustomTypes merges a slice of maps to a map
func mapCustomTypes() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if to == reflect.TypeOf(map[string]RoutineMapping{}) {
			return decodeFunctionMap(data)
		}

		if to == reflect.TypeOf(map[string]ColumnMapping{}) {
			return decodeColumnMapping(data)
		}
		if to == reflect.TypeOf(null.Bool{}) {
			return decodeNullBool(data)
		}

		return data, nil
	}
}

func decodeFunctionMap(data interface{}) (interface{}, error) {
	mappings := make(map[string]RoutineMapping)
	for dbFunctionName, functionMapping := range data.(map[string]interface{}) {

		mappedValue := new(RoutineMapping)

		// default value for map is nil pointer
		mappedValue.Model = make(map[string]ColumnMapping)
		mappedValue.Parameters = make(map[string]ParamMapping)
		mappedValue.Generate = true
		if reflect.ValueOf(functionMapping).Kind() == reflect.Bool {
			// simple processing value
			mappedValue.Generate = functionMapping.(bool)

		} else {

			err := decodeWithHook(functionMapping, mappedValue)
			if err != nil {
				return nil, fmt.Errorf("decoding complex value of FuctionMapping: %s", err)
			}

		}

		mappings[dbFunctionName] = *mappedValue

	}

	return mappings, nil
}

func decodeColumnMapping(data interface{}) (interface{}, error) {
	fromAsMap := data.(map[string]interface{})

	columns := make(map[string]ColumnMapping)
	for dbFunctionName, functionMapping := range fromAsMap {

		mappedValue := new(ColumnMapping)
		mappedValue.SelectColumn = true
		if reflect.ValueOf(functionMapping).Kind() == reflect.Bool {

			// simple processing value
			mappedValue.SelectColumn = functionMapping.(bool)

		} else {

			err := decodeWithHook(functionMapping, mappedValue)
			if err != nil {
				return nil, fmt.Errorf("decoding complex value of ColumnMapping: %s", err)
			}

			if mappedValue.MappedType == "" && mappedValue.MappingFunction != "" {
				return nil, fmt.Errorf("cant set mapping function mapping without setting mapped type")
			}
		}

		columns[dbFunctionName] = *mappedValue

	}

	return columns, nil
}

func decodeNullBool(data interface{}) (interface{}, error) {
	if reflect.TypeOf(data).Kind() == reflect.Bool {
		return null.BoolFrom(data.(bool)), nil
	}
	return null.NewBool(false, false), nil
}
