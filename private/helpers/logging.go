package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
)

const colorReset = "\033[0m"
const colorBlue = "\033[34m"
const colorRed = "\033[31m"
const colorYellow = "\033[33m"
const colorBold = "\033[1m"

func Log(msg string, args ...any) {
	log.Printf(msg, args...)
}

func LogBold(msg string, args ...any) {
	log.Printf(colorBold+msg+colorReset, args...)
}

func LogError(msg string, args ...any) {
	log.Printf(colorRed+msg+colorReset, args...)
}

func LogWarn(msg string, args ...any) {
	log.Printf(colorYellow+msg+colorReset, args...)
}

// LogDebug only log in debug in viper is set to true
func LogDebug(msg string, args ...any) {
	if viper.GetBool("debug") {
		if len(args) == 0 {
			log.Print(colorBlue + msg + colorReset)

		} else {
			log.Printf(colorBlue+msg+colorReset, args...)
		}
	}
}

func ToJson(val interface{}) string {
	formatted, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return fmt.Sprintf("error converting to json: %v", err)
	}
	return string(formatted)
}
