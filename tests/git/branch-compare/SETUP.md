## Preconditions
- The `kool` command is available in PATH
- Git is available in PATH

## Steps
1. Verify kool is available
2. Execute `kool git compare-branch <RefA> <RefB>` with optional `-C <Dir>`
3. Capture stdout, stderr, and exit code

## Context
- `kool git compare-branch` compares two git references and reports their relationship
- Three outcomes: identical (same commit), fast-forward (one is ancestor of the other), or divergent (both have unique commits)
- Supports `-C <dir>` flag to specify the git repository directory

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

type Request struct {
	RefA string
	RefB string
	Dir  string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Setup(t *testing.T, req *Request) error {
	_, err := exec.LookPath("kool")
	if err != nil {
		return fmt.Errorf("kool not found in PATH, build it first: %w", err)
	}
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := []string{"git", "compare-branch", req.RefA, req.RefB}
	if req.Dir != "" {
		args = append(args, "-C", req.Dir)
	}
	cmd := exec.Command("kool", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run kool: %w", runErr)
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```
