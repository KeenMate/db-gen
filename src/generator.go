package dbGen

import (
	"fmt"
	"os"
	"path"
	"text/template"
)

const processorsFolder = "processors"
const modelsFolder = "models"

func Generate(routines []Function, config *Config) error {
	err := ensureOutputFolder(config)
	if err != nil {
		return fmt.Errorf("ensuring output folder: %s", err)
	}
	VerboseLog("Ensured output folder")

	err = generateDbContext(routines, config)
	if err != nil {
		return fmt.Errorf("generating dbcontext: %s", err)

	}
	VerboseLog("Generated dbcontext")

	if config.GenerateModels {
		err = generateModels(routines, config)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)

		}
		VerboseLog("Generated models")

	} else {
		VerboseLog("Skipping generating models")
	}

	if config.GenerateProcessors {
		err = generateProcessors(routines, config)
		if err != nil {
			return fmt.Errorf("generating processors: %s", err)

		}
		VerboseLog("Generated processors")
	} else {
		VerboseLog("Skipping generating processors")
	}
	return nil
}

func generateDbContext(routines []Function, config *Config) error {
	dbcontextTemplate, err := parseTemplates(config.DbContextTemplate)
	if err != nil {
		return fmt.Errorf("loading dbContext template: %s", err)
	}

	data := &DbContextData{
		Functions: routines,
	}

	filepath := path.Join(config.OutputFolder, "DbConxtext.cs")
	return generateFile(data, dbcontextTemplate, filepath)
}

func generateModels(routines []Function, config *Config) error {

	moduleTemplate, err := parseTemplates(config.ModelTemplate)
	if err != nil {
		return fmt.Errorf("loading module template: %s", err)
	}

	err = os.MkdirAll(path.Join(config.OutputFolder, modelsFolder), 777)

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		filePath := path.Join(config.OutputFolder, modelsFolder, routine.ModelName+".cs")

		err = generateFile(routine, moduleTemplate, filePath)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)
		}

	}

	return nil
}

func generateProcessors(routines []Function, config *Config) error {
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
		filePath := path.Join(config.OutputFolder, processorsFolder, routine.ProcessorName+".cs")

		err = generateFile(routine, processorTemplate, filePath)
		if err != nil {
			return fmt.Errorf("generating processor %s: %s", routine.ProcessorName, err)
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

func generateFile(data interface{}, template *template.Template, filepath string) error {

	f, err := os.Create(filepath)
	defer f.Close()
	if err != nil {
		return err
	}

	err = template.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func ensureOutputFolder(config *Config) error {
	if fileExists(config.OutputFolder) {
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
