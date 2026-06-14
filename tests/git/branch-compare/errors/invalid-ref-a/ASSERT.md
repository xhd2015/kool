## Errors
- The command exits with non-zero exit code
- Stderr contains an error message about the invalid reference

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
		t.Fatalf("expected non-zero exit code for invalid ref, got 0\nstdout: %s", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "nonexistent-branch") {
		t.Fatalf("expected stderr to mention the invalid ref, got:\n%s", resp.Stderr)
	}
}
```
