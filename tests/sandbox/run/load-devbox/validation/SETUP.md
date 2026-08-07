# Scenario

**Feature**: load-devbox path validation fails hard before guest success

```
./primary --load-devbox BAD -- sh -c 'true'
  -> non-zero; stderr Error: (relative | missing | not sealed)
```

## Steps

1. Validation leaves set bad SealedLoadDevbox values; guest is trivial when reached.
2. Keep SealedDoubleDash true so flag parse still sees --load-devbox.
3. Asserts reject pre-implementation “exec --load-devbox as guest command” stderr.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Trivial guest if validation somehow passes — should not for these leaves.
	if len(req.SealedArgs) == 0 {
		req.SealedArgs = []string{"sh", "-c", "true"}
	}
	return nil
}

// isLoadDevboxExecAsCommand detects the pre-implementation failure mode where the
// sealed runner treats "--load-devbox" as the guest executable instead of a flag.
// Validation asserts must fail on this (vacuous GREEN guard).
func isLoadDevboxExecAsCommand(stderr string) bool {
	low := strings.ToLower(stderr)
	if strings.Contains(low, "executable file not found") {
		return true
	}
	if strings.Contains(low, `exec: "--load-devbox"`) || strings.Contains(stderr, `exec: "--load-devbox"`) {
		return true
	}
	if strings.Contains(low, "exec --load-devbox") {
		return true
	}
	// e.g. Error: exec --load-devbox: exec: "--load-devbox": executable file not found
	if strings.Contains(low, "exec") && strings.Contains(low, "--load-devbox") &&
		(strings.Contains(low, "not found") || strings.Contains(low, "executable")) {
		return true
	}
	return false
}
```
