package common

import (
	"os"
)

func Exit(template string, args ...any) {
	LogError(colorRed+template+colorReset, args...)
	os.Exit(1)
}

func Contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
