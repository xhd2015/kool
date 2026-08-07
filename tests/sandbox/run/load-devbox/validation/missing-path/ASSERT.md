## Expected

- Build exit 0; sealed run non-zero.
- RunStderr Error: style for a **missing load path** (path open/missing sense), after
  the runner has parsed `--load-devbox`.
- Must **not** treat `--load-devbox` as the guest command.

## Errors

- Missing absolute load-devbox path (flag parsed; target open fails).

## Exit Code

- build: 0
- sealed run: non-zero

```go
import (
	"path/filepath"
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
		t.Fatalf("expected non-zero for missing load path; stdout=%q stderr=%q",
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
	low := strings.ToLower(resp.RunStderr)
	// Missing-path sense: open failure for the path value, or load-devbox + path.
	pathHint := false
	if len(req.SealedLoadDevbox) > 0 {
		p := req.SealedLoadDevbox[0]
		if strings.Contains(resp.RunStderr, p) || strings.Contains(resp.RunStderr, filepath.Base(p)) {
			pathHint = true
		}
	}
	missingSense := strings.Contains(low, "no such file") ||
		strings.Contains(low, "not exist") ||
		strings.Contains(low, "does not exist") ||
		strings.Contains(low, "missing") ||
		// "not found" about the path — exec-as-command already rejected above.
		strings.Contains(low, "not found") ||
		strings.Contains(low, "stat ") ||
		strings.Contains(low, "open ")
	flagAndPath := (strings.Contains(low, "load-devbox") || strings.Contains(low, "load_devbox") ||
		strings.Contains(low, "devbox")) && pathHint
	if !missingSense && !flagAndPath {
		t.Fatalf("stderr must show path-open/missing failure after flag parse (path value / no such file / load-devbox+path); got %q", resp.RunStderr)
	}
}
```
