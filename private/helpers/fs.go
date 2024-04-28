package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func PathExists(path string) bool {
	// this doesnt work work if some part of path does not exist
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func FileIsReadable(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
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

func SaveAsJson(path string, data interface{}) error {
	LogDebug("Saving data as json to %s", path)

	// intent is not necessary, but will make our life easier
	saveJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("converting data to json: %s", err)
	}

	err = os.WriteFile(path, saveJson, 0777)
	if err != nil {
		return fmt.Errorf("writing file: %s", err)
	}

	return nil
}

func LoadFromJson(path string, data interface{}) error {
	LogDebug("Loading data from json file %s", path)

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %s", err)
	}

	err = json.Unmarshal(fileContent, data)
	if err != nil {
		return fmt.Errorf("parsing json: %s", err)
	}
	return nil

}
