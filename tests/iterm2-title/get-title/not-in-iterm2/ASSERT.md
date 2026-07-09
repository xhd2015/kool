## Expected

- Exit 1.
- Stderr warning mentions nothing to get and inside iTerm2.
- No AppleScript capture.
- Subcommand reserved: not open-dir `stat` of `get-title`.

## Errors

- User-facing warning; exit 1.

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
		t.Fatalf("get-title must be reserved, not open-dir:\nstderr=%q", resp.Stderr)
	}
	// Contains-only template: requirement is substring contains; CLI uses trailing \n.
	assert.Output(t, resp.Stderr, `<contains>
warning
nothing to get
inside iTerm2
</contains>
`)
	if resp.ScriptWritten || resp.CapturedScript != "" {
		t.Fatalf("expected no osascript; script written=%v content=%q", resp.ScriptWritten, resp.CapturedScript)
	}
}
```
