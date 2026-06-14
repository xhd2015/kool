## Errors
- The command exits with non-zero exit code
- Stderr contains an error message indicating the directory is not found or not a git repo

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
		t.Fatalf("expected non-zero exit code for nonexistent dir, got 0\nstdout: %s", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "no such file") && !strings.Contains(resp.Stderr, "fatal:") && !strings.Contains(resp.Stderr, "does not exist") {
		t.Fatalf("expected stderr to indicate directory not found, got:\n%s", resp.Stderr)
	}
}
```
