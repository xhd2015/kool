## Expected

- Exit 0.
- Output indicates zero tasks (0 / empty / no tasks) and/or workspace path.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("empty list exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	// Prefer explicit zero; accept empty table with workspace footer
	if !strings.Contains(out, "0") &&
		!strings.Contains(out, "empty") &&
		!strings.Contains(out, "no task") {
		// still ok if table headers only + workspace path
		if !strings.Contains(out, "label") && !strings.Contains(out, "workspace") {
			t.Fatalf("empty list should show zero/empty or headers; out:\n%s", resp.Stdout)
		}
	}
}
```
