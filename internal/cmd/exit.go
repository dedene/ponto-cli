package cmd

import "errors"

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// ExitCode returns the exit code for an error.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var ee *ExitError
	if errors.As(err, &ee) && ee != nil {
		if ee.Code < 0 {
			return 1
		}

		return ee.Code
	}

	return 1
}
