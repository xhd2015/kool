## Expected Output

```
title changed: old-session-title -> new-session-title
```

## Expected

- Exit 0.
- Stdout exactly `title changed: old-session-title -> new-session-title\n`.
- Captured script is non-empty, includes session UUID, and targets the session
  name (not only the window title).

## Side Effects

- Fake osascript invoked (script capture file written).

## Exit Code

- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
title changed: old-session-title -> new-session-title
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript for set-title success")
	}
	uuid := sessionUUID(req)
	if !strings.Contains(resp.CapturedScript, uuid) {
		t.Fatalf("script should locate session UUID %q; got:\n%s", uuid, resp.CapturedScript)
	}
	lower := strings.ToLower(resp.CapturedScript)
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should set a name property:\n%s", resp.CapturedScript)
	}
	// Default target is session/tab — reject pure window-only scripts.
	if strings.Contains(lower, "set name of current window") &&
		!strings.Contains(lower, "session") {
		t.Fatalf("session target should not be window-only:\n%s", resp.CapturedScript)
	}
}
```
