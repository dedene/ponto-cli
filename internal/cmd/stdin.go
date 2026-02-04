package cmd

import (
	"bufio"
	"os"
	"strings"
)

// ReadStdinIDs reads IDs from stdin when "-" is provided.
// Returns nil if arg is not "-".
func ReadStdinIDs(arg string) ([]string, error) {
	if arg != "-" {
		return nil, nil
	}

	var ids []string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ids = append(ids, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// IsStdin checks if the argument is stdin marker.
func IsStdin(arg string) bool {
	return arg == "-"
}
