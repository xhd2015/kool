## Expected

- Exit code non-zero
- Output mentions not exist, no such file, or cannot find

## Errors

- Command reports path does not exist

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
		t.Fatalf("expected non-zero exit code, got 0\nstdout: %s\nstderr: %s", resp.Stdout, resp.Stderr)
	}
	out := strings.ToLower(combinedOutput(resp))
	if !strings.Contains(out, "exist") && !strings.Contains(out, "no such") && !strings.Contains(out, "not found") && !strings.Contains(out, "cannot find") {
		t.Fatalf("expected error about missing path, got:\n%s", out)
	}
}
```