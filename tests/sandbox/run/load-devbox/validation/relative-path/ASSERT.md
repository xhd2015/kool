## Expected

- Build exit 0; sealed run non-zero.
- RunStderr Error: style from **flag/path validation** (absolute and/or relative wording).
- Must **not** treat `--load-devbox` as the guest command (`exec --load-devbox` /
  `executable file not found`).

## Errors

- Relative path for `--load-devbox` rejected by the runner flag parser.

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
		t.Fatalf("expected non-zero for relative --load-devbox; stdout=%q stderr=%q",
			resp.RunStdout, resp.RunStderr)
	}
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected Error: on sealed stderr")
	}
	if !strings.Contains(resp.RunStderr, "Error:") {
		t.Fatalf("stderr should use Error: prefix; got %q", resp.RunStderr)
	}
	// Vacuous GREEN guard: runner must parse --load-devbox, not exec it as guest.
	if isLoadDevboxExecAsCommand(resp.RunStderr) {
		t.Fatalf("stderr looks like guest exec of --load-devbox (flag not parsed); got %q", resp.RunStderr)
	}
	// Path validation: absolute and/or relative wording required (not flag name alone).
	low := strings.ToLower(resp.RunStderr)
	hasAbs := strings.Contains(low, "absolute")
	hasRel := strings.Contains(low, "relative")
	if !hasAbs && !hasRel {
		t.Fatalf("stderr should mention absolute and/or relative path validation; got %q", resp.RunStderr)
	}
}
```
