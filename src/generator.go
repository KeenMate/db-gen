package dbGen

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

const processorsFolder = "processors"
const modelsFolder = "models"

func Generate(routines []Function, config *Config) error {
	fileHashes, err := generateFileHashes(config.OutputFolder)
	if err != nil {
		return fmt.Errorf("generating file hashes: %s", err)
	}
	VerboseLog("Got %d file hashes", len(*fileHashes))

	err = ensureOutputFolder(config)
	if err != nil {
		return fmt.Errorf("ensuring output folder: %s", err)
	}
	log.Printf("Ensured output folder")

	err = generateDbContext(routines, fileHashes, config)
	if err != nil {
		return fmt.Errorf("generating dbcontext: %s", err)

	}
	log.Printf("Generated dbcontext")

	if config.GenerateModels {
		err = generateModels(routines, fileHashes, config)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)

		}
		log.Printf("Generated models")
	} else {
		log.Printf("Skipping generating models")
	}

	if config.GenerateProcessors {
		err = generateProcessors(routines, fileHashes, config)
		if err != nil {
			return fmt.Errorf("generating processors: %s", err)

		}
		log.Printf("Generated processors")
	} else {
		log.Printf("Skipping generating processors")
	}
	return nil
}

func generateDbContext(routines []Function, hashMap *map[string]string, config *Config) error {
	dbcontextTemplate, err := parseTemplates(config.DbContextTemplate)
	if err != nil {
		return fmt.Errorf("loading dbContext template: %s", err)
	}

	data := &DbContextData{
		Functions: routines,
	}

	fp := path.Join(config.OutputFolder, "DbContext"+config.GeneratedFileExtension)

	changed, err := generateFile(data, dbcontextTemplate, fp, hashMap)
	if err != nil {
		return err
	}

	if changed {
		VerboseLog("Updated: Dbcontext")
	} else {
		VerboseLog("Same: Dbcontext")
	}

	return nil

}

func generateModels(routines []Function, hashMap *map[string]string, config *Config) error {

	moduleTemplate, err := parseTemplates(config.ModelTemplate)
	if err != nil {
		return fmt.Errorf("loading module template: %s", err)
	}

	err = os.MkdirAll(path.Join(config.OutputFolder, modelsFolder), 777)

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		relPath := path.Join(modelsFolder, routine.ModelName+config.GeneratedFileExtension)
		filePath := path.Join(config.OutputFolder, relPath)

		changed, err := generateFile(routine, moduleTemplate, filePath, hashMap)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)
		}

		if changed {
			VerboseLog("Updated: %s", relPath)
		} else {
			VerboseLog("Same: %s", relPath)
		}
	}

	return nil
}

func generateProcessors(routines []Function, hashMap *map[string]string, config *Config) error {
	processorTemplate, err := parseTemplates(config.ProcessorTemplate)
	if err != nil {
		return fmt.Errorf("loading processor template: %s", err)
	}

	err = os.MkdirAll(path.Join(config.OutputFolder, processorsFolder), 777)
	if err != nil {
		return fmt.Errorf("creating processor output folder: %s", err)
	}

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}
		relPath := path.Join(processorsFolder, routine.ProcessorName+config.GeneratedFileExtension)
		filePath := path.Join(config.OutputFolder, relPath)

		changed, err := generateFile(routine, processorTemplate, filePath, hashMap)
		if err != nil {
			return fmt.Errorf("generating processor %s: %s", routine.ProcessorName, err)
		}

		if changed {
			VerboseLog("Updated: %s", relPath)
		} else {
			VerboseLog("Same: %s", relPath)
		}
	}

	return nil
}

func parseTemplates(filepath string) (*template.Template, error) {
	if !fileExists(filepath) {
		return nil, fmt.Errorf("config file %s does not exist", filepath)

	}

	return template.ParseFiles(filepath)
}

func generateFile(data interface{}, template *template.Template, fp string, hashMap *map[string]string) (bool, error) {

	fp = filepath.Clean(fp)

	// we want to ignore error for now
	oldHash, _ := (*hashMap)[fp]

	f, err := os.Create(fp)
	defer f.Close()
	if err != nil {
		return false, err
	}

	err = template.Execute(f, data)

	if err != nil {
		return false, err
	}

	newHash, _ := fileMd5Sum(fp)

	// only makes sense if we don't clean folder
	// otherwise we would have to keep track of what files we generated and delete the rest
	changed := newHash != oldHash

	//VerboseLog("%s -> %s", oldHash, newHash)

	return changed, nil
}

func ensureOutputFolder(config *Config) error {
	if config.ClearOutputFolder && fileExists(config.OutputFolder) {
		VerboseLog("Deleting contents of output folder")

		err := RemoveContents(config.OutputFolder)
		if err != nil {
			return fmt.Errorf("clearing output folder: %s", err)
		}
	}

	err := os.MkdirAll(config.OutputFolder, 777)
	if err != nil {
		return fmt.Errorf("creating output folder: %s", err)
	}

	return nil
}

func fileMd5Sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hashFunc := md5.New()
	if _, err := io.Copy(hashFunc, file); err != nil {
		return "", err
	}

	hash := hex.EncodeToString(hashFunc.Sum(nil))
	return hash, nil
}

func generateFileHashes(outputFolder string) (*map[string]string, error) {
	hashMap := make(map[string]string)

	err := filepath.Walk(outputFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			path = filepath.Clean(path)

			hash, err := fileMd5Sum(path)
			if err != nil {
				return err
			}

			hashMap[path] = hash
			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("generating hashes for output folder: %s", err)
	}

	return &hashMap, err

}
