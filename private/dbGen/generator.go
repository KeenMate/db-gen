package dbGen

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	common2 "github.com/keenmate/db-gen/private/helpers"
	"github.com/keenmate/db-gen/private/version"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var ValidCaseNormalized = []string{"snakecase", "camelcase", "pascalcase"}

func Generate(routines []Routine, copyTargets []CopyTarget, config *Config) error {
	fileHashes, err := generateFileHashes(config.OutputFolder)
	if err != nil {
		return fmt.Errorf("generating file hashes: %s", err)
	}

	// Also generate hashes for additional generator output folders
	for _, generator := range config.AdditionalGenerators {
		if !generator.Enabled || generator.CleanOutputFolder {
			continue
		}
		additionalHashes, err := generateFileHashes(generator.OutputFolder)
		if err != nil {
			return fmt.Errorf("generating file hashes for %s: %s", generator.Name, err)
		}
		// Merge into main hash map
		for k, v := range *additionalHashes {
			(*fileHashes)[k] = v
		}
	}

	common2.LogDebug("Got %d file hashes", len(*fileHashes))

	// Track generated files for orphan removal
	generatedFiles := make(map[string]bool)

	log.Printf("Ensuring output folder...")

	err = ensureOutputFolder(config)
	if err != nil {
		return fmt.Errorf("ensuring output folder: %s", err)
	}

	log.Printf("Generating dbcontext...")

	err = generateDbContext(routines, fileHashes, &generatedFiles, config)
	if err != nil {
		return fmt.Errorf("generating dbcontext: %s", err)

	}

	if config.GenerateModels {
		log.Printf("Generating models...")

		err = generateModels(routines, fileHashes, &generatedFiles, config)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)

		}
	} else {
		log.Printf("Skipping generating models")
	}

	if config.GenerateProcessors {
		log.Printf("Generating processors...")

		err = generateProcessors(routines, fileHashes, &generatedFiles, config)
		if err != nil {
			return fmt.Errorf("generating processors: %s", err)

		}
	} else {
		log.Printf("Skipping generating processors")
	}

	// Generate additional outputs
	err = generateAdditionalOutputs(routines, fileHashes, &generatedFiles, config)
	if err != nil {
		return fmt.Errorf("generating additional outputs: %s", err)
	}

	// Generate bulk-COPY code for configured tables
	err = generateCopyTargets(copyTargets, fileHashes, &generatedFiles, config)
	if err != nil {
		return fmt.Errorf("generating copy targets: %s", err)
	}

	// Remove orphaned files if configured
	if config.RemoveOrphanedFiles && !config.ClearOutputFolder {
		err = removeOrphanedFiles(fileHashes, &generatedFiles, config)
		if err != nil {
			return fmt.Errorf("removing orphaned files: %s", err)
		}
	}

	return nil
}

func generateDbContext(routines []Routine, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	dbContextTemplate, err := parseTemplate(config.DbContextTemplate)
	if err != nil {
		return fmt.Errorf("loading dbContext template: %s", err)
	}

	data := &DbContextData{
		Config:    config,
		Functions: routines,
		BuildInfo: version.GetBuildInfo(),
	}

	filename := changeCase("DbContext"+config.GeneratedFileExtension, config.GeneratedFileCase)
	fp := filepath.Join(config.OutputFolder, filename)

	changed, err := generateFile(data, dbContextTemplate, fp, hashMap, generatedFiles)
	if err != nil {
		return err
	}

	if changed {
		log.Printf("Updated: Dbcontext")
	} else {
		common2.LogDebug("Same: Dbcontext")
	}

	return nil

}

func generateModels(routines []Routine, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	moduleTemplate, err := parseTemplate(config.ModelTemplate)
	if err != nil {
		return fmt.Errorf("loading module template: %s", err)
	}

	err = os.MkdirAll(filepath.Join(config.OutputFolder, config.ModelsFolderName), 0777)

	for _, routine := range routines {
		if !routine.HasReturn {
			continue
		}

		filename := changeCase(routine.ModelName+config.GeneratedFileExtension, config.GeneratedFileCase)
		relPath := filepath.Join(config.ModelsFolderName, filename)
		filePath := filepath.Join(config.OutputFolder, relPath)

		data := &ModelTemplateData{
			Config:    config,
			Routine:   routine,
			BuildInfo: version.GetBuildInfo(),
		}

		changed, err := generateFile(data, moduleTemplate, filePath, hashMap, generatedFiles)
		if err != nil {
			return fmt.Errorf("generating models: %s", err)
		}

		if changed {
			log.Printf("Updated: %s", relPath)
		} else {
			common2.LogDebug("Same: %s", relPath)
		}
	}

	return nil
}

func generateProcessors(routines []Routine, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	processorTemplate, err := parseTemplate(config.ProcessorTemplate)
	if err != nil {
		return fmt.Errorf("loading processor template: %s", err)
	}

	err = os.MkdirAll(filepath.Join(config.OutputFolder, config.ProcessorsFolderName), 0777)
	if err != nil {
		return fmt.Errorf("creating processor output folder: %s", err)
	}

	for _, routine := range routines {
		// if GenerateProcessorsForVoidReturns it processors for all void returns
		if !config.GenerateProcessorsForVoidReturns && !routine.HasReturn {
			common2.LogDebug("dont generate processor for %s", routine.DbFullFunctionName)
			continue
		}

		filename := changeCase(routine.ProcessorName+config.GeneratedFileExtension, config.GeneratedFileCase)
		relPath := filepath.Join(config.ProcessorsFolderName, filename)
		filePath := filepath.Join(config.OutputFolder, relPath)

		data := &ProcessorTemplateData{
			Config:    config,
			Routine:   routine,
			BuildInfo: version.GetBuildInfo(),
		}

		changed, err := generateFile(data, processorTemplate, filePath, hashMap, generatedFiles)
		if err != nil {
			return fmt.Errorf("generating processor %s: %s", routine.ProcessorName, err)
		}

		if changed {
			log.Printf("Updated: %s", relPath)
		} else {
			common2.LogDebug("Same: %s", relPath)
		}
	}

	return nil
}

func parseTemplate(templatePath string) (*template.Template, error) {
	if !common2.PathExists(templatePath) {
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

func generateFile(data interface{}, template *template.Template, fp string, hashMap *map[string]string, generatedFiles *map[string]bool) (bool, error) {

	fp = filepath.Clean(fp)

	// Track this file as generated
	(*generatedFiles)[fp] = true

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
	if config.ClearOutputFolder && common2.PathExists(config.OutputFolder) {
		common2.LogDebug("Deleting contents of output folder")

		err := common2.RemoveContents(config.OutputFolder)
		if err != nil {
			return fmt.Errorf("clearing output folder: %s", err)
		}
	}

	err := os.MkdirAll(config.OutputFolder, 0777)
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

func generateAdditionalOutputs(routines []Routine, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	for _, generator := range config.AdditionalGenerators {
		if !generator.Enabled {
			common2.LogDebug("Skipping disabled generator: %s", generator.Name)
			continue
		}

		log.Printf("Generating %s...", generator.Name)

		// Parse template
		tmpl, err := parseTemplate(generator.Template)
		if err != nil {
			return fmt.Errorf("loading template for %s: %s", generator.Name, err)
		}

		// Clean output folder if requested
		if generator.CleanOutputFolder {
			err = os.RemoveAll(generator.OutputFolder)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("cleaning output folder for %s: %s", generator.Name, err)
			}
		}

		// Create output folder
		err = os.MkdirAll(generator.OutputFolder, 0777)
		if err != nil {
			return fmt.Errorf("creating output folder for %s: %s", generator.Name, err)
		}

		if generator.GenerationType == "single-file" {
			// Generate single file with all routines
			err = generateSingleFile(routines, tmpl, generator, hashMap, generatedFiles, config)
			if err != nil {
				return fmt.Errorf("generating %s: %s", generator.Name, err)
			}
		} else {
			// Generate one file per routine
			err = generatePerRoutineFiles(routines, tmpl, generator, hashMap, generatedFiles, config)
			if err != nil {
				return fmt.Errorf("generating %s: %s", generator.Name, err)
			}
		}
	}

	return nil
}

func generateSingleFile(routines []Routine, tmpl *template.Template, generator AdditionalGenerator, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	data := &DbContextData{
		Config:    config,
		Functions: routines,
		BuildInfo: version.GetBuildInfo(),
	}

	filePath := filepath.Join(generator.OutputFolder, generator.FileName)
	changed, err := generateFile(data, tmpl, filePath, hashMap, generatedFiles)
	if err != nil {
		return err
	}

	if changed {
		log.Printf("Updated: %s", generator.FileName)
	} else {
		common2.LogDebug("Same: %s", generator.FileName)
	}

	return nil
}

func generatePerRoutineFiles(routines []Routine, tmpl *template.Template, generator AdditionalGenerator, hashMap *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	for _, routine := range routines {
		// Skip routines without return for certain generators if needed
		if !routine.HasReturn && generator.Name == "TypeScript" {
			continue
		}

		data := &ModelTemplateData{
			Config:    config,
			Routine:   routine,
			BuildInfo: version.GetBuildInfo(),
		}

		// Determine filename
		baseName := routine.ModelName
		if generator.FileExtension != "" {
			baseName = routine.FunctionName
		}

		filename := changeCase(baseName+generator.FileExtension, generator.FileCase)
		filePath := filepath.Join(generator.OutputFolder, filename)

		changed, err := generateFile(data, tmpl, filePath, hashMap, generatedFiles)
		if err != nil {
			return fmt.Errorf("generating file for %s: %s", routine.FunctionName, err)
		}

		if changed {
			log.Printf("Updated: %s/%s", generator.Name, filename)
		} else {
			common2.LogDebug("Same: %s/%s", generator.Name, filename)
		}
	}

	return nil
}

func changeCase(str string, desiredCase string) string {
	switch desiredCase {
	case "pascalcase":
		return common2.ToPascalCase(str)
	case "camelcase":
		return common2.ToCamelCase(str)
	case "snakecase":
		return common2.ToSnakeCase(str)
	default:
		common2.LogWarn("unknown case, this should never happen")
		return str
	}
}

func removeOrphanedFiles(fileHashes *map[string]string, generatedFiles *map[string]bool, config *Config) error {
	removedCount := 0

	// Iterate through all files that existed before generation
	for filePath := range *fileHashes {
		// If the file was not regenerated, delete it
		if !(*generatedFiles)[filePath] {
			common2.LogDebug("Removing orphaned file: %s", filePath)
			err := os.Remove(filePath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("removing orphaned file %s: %s", filePath, err)
			}
			removedCount++
			log.Printf("Removed orphaned file: %s", filePath)
		}
	}

	if removedCount > 0 {
		log.Printf("Removed %d orphaned file(s)", removedCount)
	} else {
		common2.LogDebug("No orphaned files to remove")
	}

	return nil
}
