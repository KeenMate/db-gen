package dbGen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"
)

const processorsFolder = "processors"
const modelsFolder = "models"

func Generate(routines []Function, config *Config) error {

	dbcontextTemplate, err := parseTemplates("./templates/dbcontext.gotmpl")
	processorTemplate, err := parseTemplates("./templates/processor.gotmpl")
	moduleTemplate, err := parseTemplates("./templates/model.gotmpl")

	if err != nil {
		return errors.New(fmt.Sprintf("Error loading one or more templates: %s", err))
	}

	err = os.MkdirAll(config.OutputFolder, 777)
	err = os.MkdirAll(path.Join(config.OutputFolder, processorsFolder), 777)
	err = os.MkdirAll(path.Join(config.OutputFolder, modelsFolder), 777)

	if err != nil {
		return errors.New(fmt.Sprintf("Error creating output folder: %s", err))
	}

	err = generateDbContext(routines, dbcontextTemplate, config.OutputFolder)
	if err != nil {
		return errors.New(fmt.Sprintf("Error generating dbcontext: %s", err))
	}

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		err = generateModel(routine, moduleTemplate, config.OutputFolder)
		if err != nil {
			return errors.New(fmt.Sprintf("Error generating models: %s", err))
		}

		err = generateProcessor(routine, processorTemplate, config.OutputFolder)
		if err != nil {
			return errors.New(fmt.Sprintf("Error generating processors: %s", err))
		}
	}

	return nil
}

func parseTemplates(path string) (*template.Template, error) {
	return template.ParseFiles(path)
}

func generateDbContext(routines []Function, template *template.Template, outputFolder string) error {

	data := &DbContextData{
		Functions: routines,
	}

	filepath := path.Join(outputFolder, "DbConxtext.cs")
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

func generateModel(routine Function, template *template.Template, outputFolder string) error {

	filePath := path.Join(outputFolder, modelsFolder, routine.ModelName+".cs")

	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}

	err = template.Execute(f, routine)
	if err != nil {
		return err
	}

	return nil
}

func generateProcessor(routine Function, template *template.Template, outputFolder string) error {
	filePath := path.Join(outputFolder, processorsFolder, routine.ProcessorName+".cs")

	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}

	err = template.Execute(f, routine)
	if err != nil {
		return err
	}

	return nil
}
