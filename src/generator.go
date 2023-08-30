package dbGen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"
)

const basePath = "C:\\Testbench\\CSharp\\AspNetCoreDbGen\\AspNetCoreDbGen\\output"
const processorsFolder = "processors"
const modelsFolder = "models"

func Generate(routines []Function) error {

	dbcontextTemplate, err := parseTemplates("./templates/dbcontext.template")
	processorTemplate, err := parseTemplates("./templates/processor.template")
	moduleTemplate, err := parseTemplates("./templates/model.template")

	if err != nil {
		return errors.New(fmt.Sprintf("Error loading one or more templates: %s", err))
	}

	err = os.MkdirAll(basePath, 777)
	err = os.MkdirAll(path.Join(basePath, processorsFolder), 777)
	err = os.MkdirAll(path.Join(basePath, modelsFolder), 777)

	if err != nil {
		return errors.New(fmt.Sprintf("Error creating output folder: %s", err))
	}

	err = generateDbContext(routines, dbcontextTemplate)
	if err != nil {
		return errors.New(fmt.Sprintf("Error generating dbcontext: %s", err))
	}

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		err = generateModel(routine, moduleTemplate)
		if err != nil {
			return errors.New(fmt.Sprintf("Error generating models: %s", err))
		}

		err = generateProcessor(routine, processorTemplate)
		if err != nil {
			return errors.New(fmt.Sprintf("Error generating processors: %s", err))
		}
	}

	return nil
}

func parseTemplates(path string) (*template.Template, error) {
	return template.ParseFiles(path)
}

func generateDbContext(routines []Function, template *template.Template) error {

	data := &dbContextData{
		Functions: routines,
	}

	filepath := path.Join(basePath, "DbConxtext.cs")
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

func generateModel(routine Function, template *template.Template) error {

	filePath := path.Join(basePath, modelsFolder, routine.ModelName+".cs")

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

func generateProcessor(routine Function, template *template.Template) error {
	filePath := path.Join(basePath, processorsFolder, routine.ProcessorName+".cs")

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
