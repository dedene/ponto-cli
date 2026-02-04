package output

import "context"

// Mode represents the output format.
type Mode int

const (
	ModeTable Mode = iota
	ModeJSON
	ModeCSV
	ModePlain
)

type contextKey string

const modeKey contextKey = "output_mode"

// WithMode adds the output mode to the context.
func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeKey, mode)
}

// ModeFrom retrieves the output mode from the context.
func ModeFrom(ctx context.Context) Mode {
	if v, ok := ctx.Value(modeKey).(Mode); ok {
		return v
	}

	return ModeTable
}
