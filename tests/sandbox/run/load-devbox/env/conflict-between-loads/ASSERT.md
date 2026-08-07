## Expected

- Build + secondaries ok; sealed run non-zero.
- RunStderr Error: style mentioning FOO (conflict between two loads).

## Errors

- Same env key from two load sandboxes.

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
	if len(resp.SecondaryPaths) != 2 {
		t.Fatalf("expected 2 secondaries; got %v", resp.SecondaryPaths)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode == 0 {
		t.Fatalf("expected non-zero for env conflict between loads; stdout=%q stderr=%q",
			resp.RunStdout, resp.RunStderr)
	}
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected Error: on sealed stderr")
	}
	if !strings.Contains(resp.RunStderr, "Error:") && !strings.Contains(strings.ToLower(resp.RunStderr), "error") {
		t.Fatalf("stderr should be Error: style; got %q", resp.RunStderr)
	}
	if !strings.Contains(resp.RunStderr, "FOO") {
		t.Fatalf("stderr should mention FOO; got %q", resp.RunStderr)
	}
}
```
