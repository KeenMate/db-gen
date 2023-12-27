package dbGen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const tempFolder = "C:\\tmp\\db-gen"

// Saves functions to temp file on disk for development and debugging
func SaveToTempFile(content interface{}, fileNamePrefix string) (err error) {
	err = os.MkdirAll(tempFolder, 777)
	if err != nil {
		return
	}

	saveJson, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return
	}

	filename := filepath.Join(tempFolder, time.Now().Format("2006-01-02-15-04-05")+"-"+fileNamePrefix+".json")
	err = os.WriteFile(filename, saveJson, 0777)
	if err != nil {
		return
	}

	return
}

func SaveMappedFunctions() {}
