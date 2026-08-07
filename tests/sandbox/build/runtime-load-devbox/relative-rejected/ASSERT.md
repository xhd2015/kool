## Expected

- Non-zero exit.
- Stderr is non-empty and uses `Error:` style.
- Stderr mentions absolute and/or relative and/or runtime-load-devbox.

## Errors

- Relative path for `--runtime-load-devbox`.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for relative --runtime-load-devbox; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected validation error on stderr")
	}
	// Prefer Error: style (kool sandbox validation).
	if !strings.Contains(resp.Stderr, "Error:") && !strings.Contains(strings.ToLower(resp.Stderr), "error") {
		t.Fatalf("stderr should be Error: style; got %q", resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("sandbox must be routed to its handler; got %q", resp.Stderr)
	}
	// Accept any of: absolute, relative, runtime-load-devbox (requirement OR).
	hasAbs := strings.Contains(low, "absolute")
	hasRel := strings.Contains(low, "relative")
	hasFlag := strings.Contains(low, "runtime-load-devbox") ||
		strings.Contains(low, "runtime_load_devbox") ||
		strings.Contains(low, "runtime load devbox") ||
		strings.Contains(low, "load-devbox") ||
		strings.Contains(low, "load_devbox")
	if !hasAbs && !hasRel && !hasFlag {
		t.Fatalf("stderr should mention absolute/relative/runtime-load-devbox; got %q", resp.Stderr)
	}
}
```
