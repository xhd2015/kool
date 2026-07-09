## Expected

- Non-zero exit.
- Non-empty stderr (duration/usage/requires…).
- Process returns promptly (no loop).

## Errors

- Missing duration for spaced form.

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
		t.Fatalf("expected non-zero exit for missing duration; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	// Must be handled by for-every, not the top-level unknown-command router.
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("for-every must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "duration") && !strings.Contains(low, "usage") &&
		!strings.Contains(low, "require") {
		t.Fatalf("stderr should explain missing duration/usage; got %q", resp.Stderr)
	}
}
```
