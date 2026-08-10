## Expected

- Exit 0
- Stdout contains ANSI escape (`\x1b[`)
- Plan still has Would save (possibly inside color spans) and critical content
- No file written

## Errors

- None

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if strings.HasPrefix(out, "\n") {
		t.Fatalf("dry-run must not start with a leading empty line; stdout:\n%q", out)
	}
	// go-best-practice token map: green verb/kind, bold W{n}, gray meta.
	if !strings.Contains(out, "\x1b[32m") {
		t.Fatalf("expected green ANSI (Would save / grok|codex); stdout:\n%q", out)
	}
	if !strings.Contains(out, "\x1b[1m") && !strings.Contains(out, "\x1b[01m") {
		t.Fatalf("expected bold ANSI (W{n}); stdout:\n%q", out)
	}
	if !strings.Contains(out, "\x1b[90m") {
		t.Fatalf("expected gray ANSI (path/cwd/space/dry-run note); stdout:\n%q", out)
	}
	if !strings.Contains(out, "Would save") {
		t.Fatalf("missing Would save (may be colored):\n%s", out)
	}
	p := filepath.Join(req.WorkingDir, "color-plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
