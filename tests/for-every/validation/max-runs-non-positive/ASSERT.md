## Expected

- Non-zero exit.
- Stderr mentions max-runs or invalid flag value.
- Does not successfully run an infinite / unbounded loop.

## Errors

- Invalid `--max-runs` (≤ 0 when provided).

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
		t.Fatalf("expected non-zero for --max-runs 0; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("for-every must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "max-runs") && !strings.Contains(low, "max_runs") &&
		!strings.Contains(low, "invalid") && !strings.Contains(low, "greater") &&
		!strings.Contains(low, "positive") {
		t.Fatalf("stderr should reject non-positive max-runs; got %q", resp.Stderr)
	}
}
```
