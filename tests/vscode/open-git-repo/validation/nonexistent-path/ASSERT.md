## Expected
- Non-zero exit code.
- Stderr indicates path does not exist.

## Exit Code
- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for nonexistent path, got 0\nstderr: %s", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "exist") && !strings.Contains(lower, "no such") && !strings.Contains(lower, "not found") {
		t.Fatalf("expected stderr about nonexistent path, got:\n%s", resp.Stderr)
	}
}
```