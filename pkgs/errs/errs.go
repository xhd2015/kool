package errs

// ErrSilenceExitCode is an error that signals a process should exit with a specific code
// without printing an error message
type ErrSilenceExitCode struct {
	exitCode int
}

// NewSilenceExitCode creates a new ErrSilenceExitCode with the given exit code
func NewSilenceExitCode(exitCode int) *ErrSilenceExitCode {
	return &ErrSilenceExitCode{exitCode: exitCode}
}

// Error implements the error interface
func (e *ErrSilenceExitCode) Error() string {
	return ""
}

// ExitCode returns the exit code that should be used
func (e *ErrSilenceExitCode) SilenceExitCode() int {
	return e.exitCode
}

// IsSilenceExitCode checks if an error is an ErrSilenceExitCode
func IsSilenceExitCode(err error) (*ErrSilenceExitCode, bool) {
	if err == nil {
		return nil, false
	}
	e, ok := err.(*ErrSilenceExitCode)
	return e, ok
}
