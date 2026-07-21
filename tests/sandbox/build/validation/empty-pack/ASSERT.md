## Expected

- Non-zero exit.
- Stderr indicates empty pack (no files and/or no env).
- Output binary not required (may be absent or empty).

## Errors

- Empty pack after merge.

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
		t.Fatalf("expected non-zero for empty pack; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("sandbox must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "empty") && !strings.Contains(low, "no file") &&
		!strings.Contains(low, "no env") && !strings.Contains(low, "nothing") &&
		!strings.Contains(low, "require") {
		t.Fatalf("stderr should explain empty pack; got %q", resp.Stderr)
	}
}
```
