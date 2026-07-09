## Expected

- Non-zero exit.
- Stderr mentions command / usage / requires.

## Errors

- Missing command (spaced).

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
		t.Fatalf("expected non-zero for missing command; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr validation message")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("for-every must be routed to its handler; got %q", resp.Stderr)
	}
	// Require a missing-command style message (not merely the word "command").
	if !strings.Contains(low, "usage") && !strings.Contains(low, "require") &&
		!strings.Contains(low, "missing") && !strings.Contains(low, "needs") &&
		!(strings.Contains(low, "command") && (strings.Contains(low, "at least") ||
			strings.Contains(low, "provide") || strings.Contains(low, "expected") ||
			strings.Contains(low, "no command") || strings.Contains(low, "without"))) {
		t.Fatalf("stderr should explain missing child command; got %q", resp.Stderr)
	}
}
```
