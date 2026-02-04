package errfmt

import (
	"errors"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
)

// Format formats an error for display.
func Format(err error) string {
	if err == nil {
		return ""
	}

	// Check for API errors
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return formatAPIError(apiErr)
	}

	return err.Error()
}

func formatAPIError(err *api.APIError) string {
	switch err.StatusCode {
	case 401:
		return "Not authenticated. Run 'ponto auth login' first."
	case 403:
		return "Access denied. Check integration permissions in the Ponto dashboard."
	case 404:
		return fmt.Sprintf("Resource not found: %s", err.Message)
	case 429:
		return "Rate limited. Try again in a few minutes."
	default:
		if err.Message != "" {
			return fmt.Sprintf("API error (%d): %s", err.StatusCode, err.Message)
		}

		return fmt.Sprintf("API error: %d", err.StatusCode)
	}
}
