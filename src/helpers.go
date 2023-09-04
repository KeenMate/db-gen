package dbGen

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
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

// Panic Wrapper of log.panicf that adds red color
func Panic(template string, args ...any) {
	log.Panicf(colorRed+template+colorReset, args...)
}

func VerboseLog(message string) {
	if CurrentConfig.Verbose {
		log.Print(colorBlue + message + colorReset)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return false
	}

	return true
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
