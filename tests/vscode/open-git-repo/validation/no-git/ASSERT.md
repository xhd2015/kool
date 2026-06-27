## Expected
- Non-zero exit code.
- Stderr contains "not a git repository".

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
		t.Fatalf("expected non-zero exit for no-git directory, got 0\nstderr: %s", resp.Stderr)
	}
	if !strings.Contains(strings.ToLower(resp.Stderr), "not a git") {
		t.Fatalf("expected 'not a git repository' in stderr, got:\n%s", resp.Stderr)
	}
}
```