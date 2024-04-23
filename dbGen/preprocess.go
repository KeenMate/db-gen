package dbGen

import "github.com/keenmate/db-gen/common"

func PreprocessRoutines(routines *[]DbRoutine) {
	markFunctionAsOverloaded(routines)
}

func markFunctionAsOverloaded(routines *[]DbRoutine) {
	// first, we find all the overloaded functions
	namesCounter := make(map[string]int)
	for _, routine := range *routines {
		// carefull about schemas
		routineKey := routine.RoutineSchema + "." + routine.RoutineName
		count, exists := namesCounter[routineKey]

		if !exists {
			namesCounter[routineKey] = 1
		}

		namesCounter[routineKey] = count + 1
	}
	overloadedFunctionCount := 0
	// select all the value, that have overload
	for i, routine := range *routines {
		// carefull about schemas
		routineKey := routine.RoutineSchema + "." + routine.RoutineName
		count, exists := namesCounter[routineKey]

		if !exists || count == 1 {
			continue
		}

		overloadedFunctionCount++
		(*routines)[i].HasOverload = true

	}

	common.Log("Marked %d functions as overload", overloadedFunctionCount)

}
