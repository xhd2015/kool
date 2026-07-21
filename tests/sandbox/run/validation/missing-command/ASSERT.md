## Expected

- Build succeeds (sealed binary exists).
- Sealed binary was executed (`RunExecuted`).
- Sealed run exit code is non-zero.
- Sealed stderr is non-empty and mentions command and/or usage (not merely a
  generic “not implemented” stub without guidance).

## Errors

- Missing guest command / usage.

## Exit Code

- build: 0
- sealed run: non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists {
		t.Fatalf("expected sealed binary at %q", resp.OutputPath)
	}
	if !resp.RunExecuted {
		t.Fatal("expected AfterBuildRun to execute sealed binary")
	}
	if resp.RunExitCode == 0 {
		t.Fatalf("expected non-zero sealed exit for missing command; stdout=%q stderr=%q", resp.RunStdout, resp.RunStderr)
	}
	low := strings.ToLower(resp.RunStderr)
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected Error: style message on sealed stderr")
	}
	// Require command/usage guidance — stub "Error: run not implemented" stays RED.
	if !strings.Contains(low, "command") && !strings.Contains(low, "usage") {
		t.Fatalf("stderr should mention command/usage; got %q", resp.RunStderr)
	}
	if !strings.Contains(resp.RunStderr, "Error:") && !strings.Contains(low, "error") {
		t.Fatalf("stderr should be Error: style; got %q", resp.RunStderr)
	}
}
```
