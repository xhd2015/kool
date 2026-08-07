## Expected

- Build succeeds; sealed run non-zero.
- RunStderr is Error: style and mentions FOO and conflict/incompatible (or both sources:
  current sandbox vs load path).

## Errors

- Same env key from two sandboxes (primary + load).

## Exit Code

- build: 0
- sealed run: non-zero

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
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode == 0 {
		t.Fatalf("expected non-zero sealed exit for env conflict; stdout=%q stderr=%q",
			resp.RunStdout, resp.RunStderr)
	}
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected Error: on sealed stderr")
	}
	if !strings.Contains(resp.RunStderr, "Error:") && !strings.Contains(strings.ToLower(resp.RunStderr), "error") {
		t.Fatalf("stderr should be Error: style; got %q", resp.RunStderr)
	}
	if !strings.Contains(resp.RunStderr, "FOO") {
		t.Fatalf("stderr should mention conflicting key FOO; got %q", resp.RunStderr)
	}
	low := strings.ToLower(resp.RunStderr)
	hasConflict := strings.Contains(low, "incompatible") ||
		strings.Contains(low, "conflict") ||
		strings.Contains(low, "current sandbox") ||
		(strings.Contains(low, "sandbox") && strings.Contains(low, "load"))
	if !hasConflict {
		t.Fatalf("stderr should mention incompatible/conflict or sources; got %q", resp.RunStderr)
	}
}
```
