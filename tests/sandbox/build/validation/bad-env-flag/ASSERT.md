## Expected

- Non-zero exit.
- Stderr mentions env and/or `=`.

## Errors

- Invalid `--env` form (no `=`).

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
		t.Fatalf("expected non-zero for bad --env; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("sandbox must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "env") && !strings.Contains(resp.Stderr, "=") {
		t.Fatalf("stderr should mention env or =; got %q", resp.Stderr)
	}
}
```
