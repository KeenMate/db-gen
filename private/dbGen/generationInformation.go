package dbGen

import (
	"bufio"
	"fmt"
	"github.com/keenmate/db-gen/private/helpers"
	"github.com/keenmate/db-gen/private/version"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Stores information from last generation, used for getting diff

const generationInfoFileName = "db-gen_generation_info.json"

type GenerationInformation struct {
	Version  string      `json:"version"`
	Time     time.Time   `json:"time"`
	Routines []DbRoutine `json:"routines"`
}

type databaseChanges struct {
	deletedRoutines []DbRoutine
	createdRoutines []DbRoutine
	changedRoutines []routineChanges
}

type routineChanges struct {
	oldRoutine DbRoutine
	newRoutine DbRoutine
}

// LoadGenerationInformation if there is error reading, just behave like there is not info
func LoadGenerationInformation(config *Config) (*GenerationInformation, bool) {
	path := filepath.Join(config.OutputFolder, generationInfoFileName)
	if !helpers.FileIsReadable(path) {
		return nil, false
	}

	info := &GenerationInformation{}

	err := helpers.LoadFromJson(path, info)
	if err != nil {
		return nil, false
	}

	return info, true
}

func SaveGenerationInformation(config *Config, routines []DbRoutine, version string) error {
	info := GenerationInformation{
		Version:  version,
		Time:     time.Now(),
		Routines: routines,
	}

	path := filepath.Join(config.OutputFolder, generationInfoFileName)
	err := helpers.SaveAsJson(path, info)
	if err != nil {
		return fmt.Errorf("saving generation information: %w", err)
	}

	return nil
}

// CheckVersion if version is not same, ask user to confirm and return false if he cancels generation
func (info *GenerationInformation) CheckVersion() bool {

	if info.Version == version.GetVersion() {
		return true
	}

	// ask user for confirmation
	message := fmt.Sprintf("Db-gen version changed from %s to %s, are you sure you want to continue", info.Version, version.GetVersion())
	if !askForConfirmation(message) {
		return false
	}

	return true
}

func (info *GenerationInformation) GetRoutinesChanges(newRoutines []DbRoutine) string {
	oldRoutines := info.Routines
	changes := new(databaseChanges)

	// handle deleted routines
	for _, oldRoutine := range oldRoutines {
		_, exists := findRoutine(newRoutines, oldRoutine)
		if exists {
			continue
		}
		changes.deletedRoutines = append(changes.deletedRoutines, oldRoutine)
	}

	for _, newRoutine := range newRoutines {
		oldRoutine, exists := findRoutine(oldRoutines, newRoutine)
		if !exists {
			changes.createdRoutines = append(changes.deletedRoutines, newRoutine)
			continue
		}

		changes.changedRoutines = append(changes.changedRoutines, routineChanges{
			oldRoutine: *oldRoutine,
			newRoutine: newRoutine,
		})
	}

	return changes.String()
}

func (databaseChanges *databaseChanges) String() string {
	var out strings.Builder

	if len(databaseChanges.deletedRoutines) > 0 {
		out.WriteString("Deleted routines:\n")
		for _, routine := range databaseChanges.deletedRoutines {
			out.WriteString(fmt.Sprintf(" - %s.%s\n", routine.RoutineSchema, routine.RoutineName))
		}
	}

	if len(databaseChanges.createdRoutines) > 0 {
		out.WriteString("Created routines:\n")
		for _, routine := range databaseChanges.createdRoutines {
			out.WriteString(fmt.Sprintf(" - %s.%s\n", routine.RoutineSchema, routine.RoutineName))
		}
	}

	if len(databaseChanges.changedRoutines) > 0 {
		out.WriteString("Changed routines:\n")
		for _, changes := range databaseChanges.changedRoutines {
			out.WriteString(changes.String())
		}
	}

	return out.String()
}

func (change *routineChanges) String() string {
	var changes strings.Builder

	oldR := change.oldRoutine
	newR := change.newRoutine

	if oldR.FuncType != newR.FuncType {
		changes.WriteString(fmt.Sprintf("\t Function type changed from %s to %s \n", oldR.FuncType, newR.FuncType))
	}

	// parameters
	parametersChanges := getParameterChanges(oldR.InParameters, newR.InParameters)
	if len(parametersChanges) > 0 {
		changes.WriteString(parametersChanges)
	}

	// model
	modelChanges := getParameterChanges(oldR.OutParameters, newR.OutParameters)
	if len(modelChanges) > 0 {
		changes.WriteString(modelChanges)
	}

	changesString := changes.String()
	// TODO it would be ideal to separate detecting if routines changed
	if len(changesString) == 0 {
		return ""
	}

	return fmt.Sprintf(" - %s.%s: \n %s", newR.RoutineSchema, newR.RoutineName, changesString)
}

func getParameterChanges(oldParams []DbParameter, newParams []DbParameter) string {
	var outBuilder strings.Builder

	slices.SortFunc(oldParams, func(a, b DbParameter) int {
		return a.OrdinalPosition - b.OrdinalPosition
	})
	slices.SortFunc(newParams, func(a, b DbParameter) int {
		return a.OrdinalPosition - b.OrdinalPosition
	})

	oldLength := len(oldParams)
	newLength := len(newParams)

	for i := 0; i < min(oldLength, newLength); i++ {
		oldParam := oldParams[i]
		newParam := newParams[i]

		// handle changes
		if oldParam.Name != newParam.Name {
			outBuilder.WriteString(fmt.Sprintf("\t parameter %d(%s): renamed from %s\n", i, newParam.Name, oldParam.Name))
		}

		if oldParam.IsNullable != newParam.IsNullable {
			outBuilder.WriteString(fmt.Sprintf("\t parameter %d(%s): nulability changed from %v to %v\n", i, newParam.Name, oldParam.IsNullable, newParam.IsNullable))
		}

		if oldParam.IsOptional != newParam.IsOptional {
			outBuilder.WriteString(fmt.Sprintf("\t parameter %d(%s): optional changed from %v to %v\n", i, newParam.Name, oldParam.IsOptional, newParam.IsOptional))
		}

		if oldParam.UDTName != newParam.UDTName {
			outBuilder.WriteString(fmt.Sprintf("\t parameter %d(%s): data type changed from %v to %v\n", i, newParam.Name, oldParam.UDTName, newParam.UDTName))
		}

	}

	if oldLength > newLength {
		for _, oldParam := range oldParams[newLength:oldLength] {
			outBuilder.WriteString(fmt.Sprintf("\t removed parameter: %s\n", oldParam.Name))
		}
	}

	if newLength > oldLength {
		for _, oldParam := range oldParams[oldLength:newLength] {
			outBuilder.WriteString(fmt.Sprintf("\t added parameter: %s\n", oldParam.Name))
		}
	}

	return outBuilder.String()
}

func findRoutine(routines []DbRoutine, routine DbRoutine) (*DbRoutine, bool) {
	index := slices.IndexFunc(routines, func(oldRoutine DbRoutine) bool {
		return routine.RoutineName == oldRoutine.RoutineName && routine.RoutineSchema == oldRoutine.RoutineSchema
	})

	if index == -1 {
		return nil, false
	}

	return &routines[index], true
}

func askForConfirmation(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/n]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true
	}
	return false

}
