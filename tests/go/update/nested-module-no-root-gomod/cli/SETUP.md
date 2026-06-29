# Scenario

**Feature**: `kool go update` CLI for nested module without root go.mod

## Steps

1. Verify kool is in PATH
2. Build dot-pkgs-like fixture
3. Set operation to CLI update

```go
import (
	"fmt"
	"os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("kool"); err != nil {
		return fmt.Errorf("kool not found in PATH, build it first: %w", err)
	}
	req.Operation = "cli"
	return nil
}
```