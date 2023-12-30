package dbGen

import (
	"encoding/json"
	"log"
	"os"
)

func PrettyPrintSlice[T interface{}](values []T) {
	for i, val := range values {
		formattedStr, _ := json.MarshalIndent(val, "", "  ")
		log.Printf("%d.	%s", i, formattedStr)
	}
}
func PrettyPrint(val interface{}) {
	formattedStr, _ := json.MarshalIndent(val, "", "  ")
	log.Printf("%s", formattedStr)

}

const colorReset = "\033[0m"
const colorBlue = "\033[34m"
const colorRed = "\033[31m"

func Exit(template string, args ...any) {
	log.Printf(colorRed+template+colorReset, args...)
	os.Exit(1)
}

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
