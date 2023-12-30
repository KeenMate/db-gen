package dbGen

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/keenmate/db-gen/common"
	"github.com/stoewer/go-strcase"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const processorsFolder = "processors"
const modelsFolder = "models"

func Generate(routines []Routine, config *Config) error {
	fileHashes, err := generateFileHashes(config.OutputFolder)
	if err != nil {
		return fmt.Errorf("generating file hashes: %s", err)
	}
	common.LogDebug("Got %d file hashes", len(*fileHashes))

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

func generateDbContext(routines []Routine, hashMap *map[string]string, config *Config) error {
	dbcontextTemplate, err := parseTemplates(config.DbContextTemplate)
	if err != nil {
		return fmt.Errorf("loading dbContext template: %s", err)
	}

	data := &DbContextData{
		Functions: routines,
	}

	filename := changeCase("DbContext"+config.GeneratedFileExtension, config.GeneratedFileCase)
	fp := filepath.Join(config.OutputFolder, filename)

	changed, err := generateFile(data, dbcontextTemplate, fp, hashMap)
	if err != nil {
		return err
	}

	if changed {
		log.Printf("Updated: Dbcontext")
	} else {
		common.LogDebug("Same: Dbcontext")
	}

	return nil

}

func generateModels(routines []Routine, hashMap *map[string]string, config *Config) error {

	moduleTemplate, err := parseTemplates(config.ModelTemplate)
	if err != nil {
		return fmt.Errorf("loading module template: %s", err)
	}

	err = os.MkdirAll(filepath.Join(config.OutputFolder, modelsFolder), 777)

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		filename := changeCase(routine.ModelName+config.GeneratedFileExtension, config.GeneratedFileCase)
		relPath := filepath.Join(modelsFolder, filename)
		filePath := filepath.Join(config.OutputFolder, relPath)

		changed, err := generateFile(routine, moduleTemplate, filePath, hashMap)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)
		}

		if changed {
			log.Printf("Updated: %s", relPath)
		} else {
			common.LogDebug("Same: %s", relPath)
		}
	}

	return nil
}

func generateProcessors(routines []Routine, hashMap *map[string]string, config *Config) error {
	processorTemplate, err := parseTemplates(config.ProcessorTemplate)
	if err != nil {
		return fmt.Errorf("loading processor template: %s", err)
	}

	err = os.MkdirAll(filepath.Join(config.OutputFolder, processorsFolder), 777)
	if err != nil {
		return fmt.Errorf("creating processor output folder: %s", err)
	}

	for _, routine := range routines {
		// if GenerateProcessorsForVoidReturns it processors for all void returns
		if !config.GenerateProcessorsForVoidReturns && !routine.HasReturn {
			common.LogDebug("dont generate processor for %s", routine.DbFullFunctionName)
			continue
		}

		filename := changeCase(routine.ProcessorName+config.GeneratedFileExtension, config.GeneratedFileCase)
		relPath := filepath.Join(processorsFolder, filename)
		filePath := filepath.Join(config.OutputFolder, relPath)

		changed, err := generateFile(routine, processorTemplate, filePath, hashMap)
		if err != nil {
			return fmt.Errorf("generating processor %s: %s", routine.ProcessorName, err)
		}

		if changed {
			log.Printf("Updated: %s", relPath)
		} else {
			common.LogDebug("Same: %s", relPath)
		}
	}

	return nil
}

func parseTemplates(templatePath string) (*template.Template, error) {
	if !common.PathExists(templatePath) {
		return nil, fmt.Errorf("template file %s does not exist", templatePath)

	}

	name := filepath.Base(templatePath)

	tmpl, err := template.New(name).
		Funcs(getTemplateFunctions()).
		ParseFiles(templatePath)

	if err != nil {
		return nil, err
	}
	return tmpl, nil
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

	//DebugLog("%s -> %s", oldHash, newHash)

	return changed, nil
}

func ensureOutputFolder(config *Config) error {
	if config.ClearOutputFolder && common.PathExists(config.OutputFolder) {
		common.LogDebug("Deleting contents of output folder")

		err := common.RemoveContents(config.OutputFolder)
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

	// If folder doesnt exist
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		return &hashMap, nil
	}

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

var ValidCase = []string{"snake", "lcamel", "ucamel"}

func changeCase(str string, desiredCase string) string {
	switch strings.ToLower(desiredCase) {
	case "ucamel":
		return strcase.UpperCamelCase(str)
	case "lcamel":
		return strcase.LowerCamelCase(str)
	case "snake":
		return strcase.SnakeCase(str)
	default:

		return str
	}
}
