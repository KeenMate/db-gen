package dbGen

import (
	"encoding/json"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"log"
	"os"
	"path/filepath"
	"time"
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

// Exit Wrapper of log.panicf that adds red color
func Exit(template string, args ...any) {
	log.Printf(colorRed+template+colorReset, args...)
	os.Exit(1)
}

func VerboseLog(message string, args ...any) {
	if CurrentConfig.Verbose {
		if len(args) == 0 {
			log.Print(colorBlue + message + colorReset)

		} else {
			log.Printf(colorBlue+message+colorReset, args...)
		}
	}
}

func VerboseLogStruct(val interface{}) {
	if CurrentConfig.Verbose {
		formattedStr, _ := json.MarshalIndent(val, "", "  ")
		log.Printf(colorBlue+"%s"+colorReset, formattedStr)
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

const bold = "\033[5;1m"

// Hello prints welcome message and waits fo a bit
func Hello() {
	figure.NewColorFigure("db-gen", "", "green", true).Print()
	fmt.Println("Ultimate db call code generator by " + bold + "KEEN|MATE" + colorReset)
	fmt.Println()

	time.Sleep(2 * time.Second)

}
