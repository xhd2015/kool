## Expected

- Exit code 0.
- Stdout is root `iterm2` help (not open-dir-only): mentions **`sessions`**,
  **`snapshot`**, and **`status`** (session status surface).
- Prefer also mentioning the singular **`session`** command word (locked: require
  `session` so `session <id> status` is indexed, not only the word "status").
- Stdout ends with a trailing newline.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	// Root help must index both multi-session snapshot and single-session status.
	for _, want := range []string{"sessions", "snapshot", "session", "status"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root iterm2 help missing %q:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", out)
	}
}
```
