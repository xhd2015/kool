## Expected

- Non-zero exit.
- Stderr mentions output / `-o` / `--output`.
- Must be handled by sandbox (not top-level unrecognized command only, once wired).

## Errors

- Missing `-o` / `--output`.

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
		t.Fatalf("expected non-zero for missing -o; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("sandbox must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "output") && !strings.Contains(low, "-o") {
		t.Fatalf("stderr should mention missing output/-o; got %q", resp.Stderr)
	}
}
```
