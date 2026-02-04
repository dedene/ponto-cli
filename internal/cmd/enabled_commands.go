package cmd

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

func enforceEnabledCommands(kctx *kong.Context, enabled string) error {
	enabled = strings.TrimSpace(enabled)
	if enabled == "" {
		return nil
	}

	allow := parseEnabledCommands(enabled)
	if len(allow) == 0 {
		return nil
	}

	if allow["*"] || allow["all"] {
		return nil
	}

	// Get full command path (e.g., "accounts list" -> "accounts.list")
	cmdPath := strings.ReplaceAll(kctx.Command(), " ", ".")
	cmdPath = strings.ToLower(cmdPath)

	// Check exact match first
	if allow[cmdPath] {
		return nil
	}

	// Check top-level command
	parts := strings.Split(cmdPath, ".")
	if len(parts) > 0 && allow[parts[0]] {
		return nil
	}

	return &ExitError{
		Code: 2,
		Err:  fmt.Errorf("command %q is not enabled (allowed: %s)", cmdPath, enabled),
	}
}

func parseEnabledCommands(value string) map[string]bool {
	out := map[string]bool{}

	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}

		out[part] = true
	}

	return out
}
