## Expected

- Exit 1.
- Stderr contains a warning that nothing can be set and the command must run
  inside iTerm2 (wording may include `nothing to set` and `inside iTerm2`).
- No AppleScript is written (fake osascript never invoked).
- Subcommand is reserved: not an open-dir `stat` error for `set-title`.

## Errors

- User-facing warning on stderr; silent non-zero exit.

## Exit Code

- 1

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("exit=%d want 1\nstdout=%q\nstderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "stat") || strings.Contains(low, "no such file") {
		t.Fatalf("set-title must be reserved, not open-dir:\nstderr=%q", resp.Stderr)
	}
	// Contains-only template: requirement is substring contains; CLI uses trailing \n.
	assert.Output(t, resp.Stderr, `<contains>
warning
nothing to set
inside iTerm2
</contains>
`)
	if resp.ScriptWritten || resp.CapturedScript != "" {
		t.Fatalf("expected no osascript; script written=%v content=%q", resp.ScriptWritten, resp.CapturedScript)
	}
	if strings.Contains(resp.Stdout, "title changed:") {
		t.Fatalf("unexpected success stdout: %q", resp.Stdout)
	}
}
```
