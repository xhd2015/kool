## Expected
- Non-zero exit code.
- Stderr indicates path is not a directory.

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
		t.Fatalf("expected non-zero exit for file path, got 0\nstderr: %s", resp.Stderr)
	}
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "directory") && !strings.Contains(lower, "not a dir") {
		t.Fatalf("expected stderr about not a directory, got:\n%s", resp.Stderr)
	}
}
```