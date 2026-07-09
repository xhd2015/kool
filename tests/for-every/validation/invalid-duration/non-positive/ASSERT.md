## Expected

- Non-zero exit.
- Stderr indicates duration must be greater than 0 (or invalid / non-positive).

## Errors

- Duration ≤ 0.

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
		t.Fatalf("expected non-zero for 0s duration; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("for-every must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "duration") && !strings.Contains(low, "greater") &&
		!strings.Contains(low, "positive") {
		t.Fatalf("stderr should reject non-positive duration; got %q", resp.Stderr)
	}
}
```
