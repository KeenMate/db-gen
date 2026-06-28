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
	Tables   []DbTable   `json:"tables"`
}

type databaseChanges struct {
	deletedRoutines      []DbRoutine
	createdRoutines      []DbRoutine
	maybeChangedRoutines []routinePain
}

type routinePain struct {
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

// SaveGenerationInformation Saves json with generation info inside output folder.
func SaveGenerationInformation(config *Config, routines []DbRoutine, tables []DbTable, version string) error {
	// make local copy because we will be removing specific name
	routinesCopy := make([]DbRoutine, len(routines))
	copy(routinesCopy, routines)

	for i := range routinesCopy {
		routinesCopy[i].SpecificName = ""
	}

	info := GenerationInformation{
		Version:  version,
		Time:     time.Now(),
		Routines: routinesCopy,
		Tables:   tables,
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
			changes.createdRoutines = append(changes.createdRoutines, newRoutine)
			continue
		}

		changes.maybeChangedRoutines = append(changes.maybeChangedRoutines, routinePain{
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
			out.WriteString(fmt.Sprintf(" - %s.%s\n", routine.RoutineSchema, routine.RoutineNameWithParams))
		}
	}

	if len(databaseChanges.createdRoutines) > 0 {
		out.WriteString("Created routines:\n")
		for _, routine := range databaseChanges.createdRoutines {
			out.WriteString(fmt.Sprintf(" - %s.%s\n", routine.RoutineSchema, routine.RoutineNameWithParams))
		}
	}

	log.Printf("number of changed routines: %d\n", len(databaseChanges.maybeChangedRoutines))
	if len(databaseChanges.maybeChangedRoutines) > 0 {
		changesDetected := false
		for _, routinesInfo := range databaseChanges.maybeChangedRoutines {
			changes := routinesInfo.String()
			if !changesDetected && len(changes) > 0 {
				out.WriteString("Changed routines:\n")
				changesDetected = true
			}
			out.WriteString(changes)

		}
	}

	return out.String()
}

func (change *routinePain) String() string {
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
		for _, oldParam := range newParams[oldLength:newLength] {
			outBuilder.WriteString(fmt.Sprintf("\t added parameter: %s\n", oldParam.Name))
		}
	}

	return outBuilder.String()
}

// ---------------------------------------------------------------------------
// Copy-target (table) change detection
// ---------------------------------------------------------------------------

type tableChanges struct {
	deletedTables []DbTable
	createdTables []DbTable
	changedTables []tablePair
}

type tablePair struct {
	oldTable DbTable
	newTable DbTable
}

// GetTableChanges diffs the configured copy-target tables against the last generation.
func (info *GenerationInformation) GetTableChanges(newTables []DbTable) string {
	oldTables := info.Tables
	changes := new(tableChanges)

	for _, oldTable := range oldTables {
		if _, exists := findTable(newTables, oldTable); !exists {
			changes.deletedTables = append(changes.deletedTables, oldTable)
		}
	}

	for _, newTable := range newTables {
		oldTable, exists := findTable(oldTables, newTable)
		if !exists {
			changes.createdTables = append(changes.createdTables, newTable)
			continue
		}

		changes.changedTables = append(changes.changedTables, tablePair{
			oldTable: *oldTable,
			newTable: newTable,
		})
	}

	return changes.String()
}

func (tableChanges *tableChanges) String() string {
	var out strings.Builder

	if len(tableChanges.deletedTables) > 0 {
		out.WriteString("Deleted copy-target tables:\n")
		for _, table := range tableChanges.deletedTables {
			out.WriteString(fmt.Sprintf(" - %s.%s\n", table.TableSchema, table.TableName))
		}
	}

	if len(tableChanges.createdTables) > 0 {
		out.WriteString("Created copy-target tables:\n")
		for _, table := range tableChanges.createdTables {
			out.WriteString(fmt.Sprintf(" - %s.%s\n", table.TableSchema, table.TableName))
		}
	}

	changesDetected := false
	for _, pair := range tableChanges.changedTables {
		columnChanges := getColumnChanges(pair.oldTable.Columns, pair.newTable.Columns)
		if len(columnChanges) == 0 {
			continue
		}
		if !changesDetected {
			out.WriteString("Changed copy-target tables:\n")
			changesDetected = true
		}
		out.WriteString(fmt.Sprintf(" - %s.%s: \n%s", pair.newTable.TableSchema, pair.newTable.TableName, columnChanges))
	}

	return out.String()
}

func getColumnChanges(oldColumns []DbColumn, newColumns []DbColumn) string {
	var out strings.Builder

	slices.SortFunc(oldColumns, func(a, b DbColumn) int { return a.OrdinalPosition - b.OrdinalPosition })
	slices.SortFunc(newColumns, func(a, b DbColumn) int { return a.OrdinalPosition - b.OrdinalPosition })

	oldLength := len(oldColumns)
	newLength := len(newColumns)

	for i := 0; i < min(oldLength, newLength); i++ {
		oldColumn := oldColumns[i]
		newColumn := newColumns[i]

		if oldColumn.Name != newColumn.Name {
			out.WriteString(fmt.Sprintf("\t column %d(%s): renamed from %s\n", i, newColumn.Name, oldColumn.Name))
		}

		if oldColumn.UDTName != newColumn.UDTName {
			out.WriteString(fmt.Sprintf("\t column %d(%s): data type changed from %s to %s\n", i, newColumn.Name, oldColumn.UDTName, newColumn.UDTName))
		}

		if oldColumn.IsNullable != newColumn.IsNullable {
			out.WriteString(fmt.Sprintf("\t column %d(%s): nullability changed from %v to %v\n", i, newColumn.Name, oldColumn.IsNullable, newColumn.IsNullable))
		}

		if oldColumn.HasDefault != newColumn.HasDefault {
			out.WriteString(fmt.Sprintf("\t column %d(%s): default changed from %v to %v\n", i, newColumn.Name, oldColumn.HasDefault, newColumn.HasDefault))
		}
	}

	if oldLength > newLength {
		for _, oldColumn := range oldColumns[newLength:oldLength] {
			out.WriteString(fmt.Sprintf("\t removed column: %s\n", oldColumn.Name))
		}
	}

	if newLength > oldLength {
		for _, newColumn := range newColumns[oldLength:newLength] {
			out.WriteString(fmt.Sprintf("\t added column: %s\n", newColumn.Name))
		}
	}

	return out.String()
}

func findTable(tables []DbTable, table DbTable) (*DbTable, bool) {
	index := slices.IndexFunc(tables, func(t DbTable) bool {
		return t.TableSchema == table.TableSchema && t.TableName == table.TableName
	})

	if index == -1 {
		return nil, false
	}

	return &tables[index], true
}

func findRoutine(routines []DbRoutine, routine DbRoutine) (*DbRoutine, bool) {
	index := slices.IndexFunc(routines, func(oldRoutine DbRoutine) bool {
		// for function with overloads we can't detect parameter changes
		if routine.HasOverload {
			return routine.RoutineNameWithParams == oldRoutine.RoutineNameWithParams
		}

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
