## Expected

- Build succeeds (flag + pack accepted); sealed binary executed.
- Sealed run exit is non-zero.
- Sealed stderr is non-empty and mentions `HOME` and/or `home-linked` /
  `home linked` (policy error, not silent guest success).

## Errors

- Packed HOME cannot be set when `--home-linked` (value ≠ abs SANDBOX_ROOT).

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
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists {
		t.Fatalf("expected sealed binary at %q", resp.OutputPath)
	}
	if !resp.RunExecuted {
		t.Fatal("expected AfterBuildRun to execute sealed binary")
	}
	if resp.RunExitCode == 0 {
		t.Fatalf("expected non-zero sealed exit for packed HOME conflict; stdout=%q stderr=%q",
			resp.RunStdout, resp.RunStderr)
	}
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected policy error on sealed stderr")
	}
	low := strings.ToLower(resp.RunStderr)
	hasHome := strings.Contains(low, "home")
	hasLinked := strings.Contains(low, "home-linked") || strings.Contains(low, "home linked") || strings.Contains(low, "homelinked")
	if !hasHome && !hasLinked {
		t.Fatalf("stderr should mention HOME and/or home-linked; got %q", resp.RunStderr)
	}
}
```
