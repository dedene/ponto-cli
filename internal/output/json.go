package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSON outputs data as JSON.
func JSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// JSONCompact outputs data as compact JSON.
func JSONCompact(v any) error {
	enc := json.NewEncoder(os.Stdout)

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}
