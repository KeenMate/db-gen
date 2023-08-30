package dbGen

import (
	"encoding/json"
	"log"
)

func PrettyPrint[T interface{}](values []T) {
	for i, val := range values {
		formattedStr, _ := json.MarshalIndent(val, "", "  ")
		log.Printf("%d.	%s", i, formattedStr)
	}
}
