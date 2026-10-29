package debug

import (
	"encoding/json"
	"log"
)

func LogJSON[T any](label string, value T) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] %s: JSON marshal error: %v", label, err)
		return
	}

	log.Printf("[DEBUG] %s:\n%s", label, data)
}
