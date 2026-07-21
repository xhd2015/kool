## Expected

- Non-zero exit.
- Stderr indicates missing local file / not found / no such file.

## Errors

- Missing local path for `--file`.

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
		t.Fatalf("expected non-zero for missing --file source; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("sandbox must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "missing") && !strings.Contains(low, "not found") &&
		!strings.Contains(low, "no such") && !strings.Contains(low, "exist") &&
		!strings.Contains(low, "file") {
		t.Fatalf("stderr should explain missing local file; got %q", resp.Stderr)
	}
}
```
