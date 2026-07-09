## Expected Output

```
title changed: old-window-title -> new-window-title
```

## Expected

- Exit 0; success stdout with trailing newline.
- Captured script includes session UUID and operates on a **window** name
  (contains `window` near name assignment).

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
title changed: old-window-title -> new-window-title
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript")
	}
	uuid := sessionUUID(req)
	if !strings.Contains(resp.CapturedScript, uuid) {
		t.Fatalf("script should locate session UUID %q; got:\n%s", uuid, resp.CapturedScript)
	}
	lower := strings.ToLower(resp.CapturedScript)
	if !strings.Contains(lower, "window") {
		t.Fatalf("--window script should reference window:\n%s", resp.CapturedScript)
	}
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should set a name property:\n%s", resp.CapturedScript)
	}
}
```
