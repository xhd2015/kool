# Scenario

**Feature**: `kool go replace` CLI for nested module without root go.mod

## Steps

1. Verify kool is in PATH
2. Build dot-pkgs-like fixture
3. Set operation to CLI replace

```go
import (
	"fmt"
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := exec.LookPath("kool"); err != nil {
		return fmt.Errorf("kool not found in PATH, build it first: %w", err)
	}
	req.Operation = "cli"
	return nil
}
```