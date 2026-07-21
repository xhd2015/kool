## Expected

- Non-zero exit.
- Stderr indicates unknown / unrecognized / invalid command (or names nosuch).
- StartSession not called.

## Errors

- Unknown subcommand.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for unknown subcommand; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr for unknown subcommand")
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run for unknown subcommand")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "unknown") && !strings.Contains(low, "unrecognized") &&
		!strings.Contains(low, "invalid") && !strings.Contains(low, "nosuch") &&
		!strings.Contains(low, "not found") {
		t.Fatalf("stderr should indicate unknown subcommand; got %q", resp.Stderr)
	}
}
```
