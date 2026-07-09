## Expected Output

```
Tab Alpha
```

## Expected

- Exit 0.
- Stdout is exactly the title plus trailing newline.
- Captured script non-empty; includes session UUID; targets session name.

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
Tab Alpha
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript for get-title")
	}
	uuid := sessionUUID(req)
	if !strings.Contains(resp.CapturedScript, uuid) {
		t.Fatalf("script should locate session UUID %q; got:\n%s", uuid, resp.CapturedScript)
	}
	lower := strings.ToLower(resp.CapturedScript)
	if !strings.Contains(lower, "name") {
		t.Fatalf("script should read a name property:\n%s", resp.CapturedScript)
	}
	if strings.Contains(lower, "name of current window") && !strings.Contains(lower, "session") {
		t.Fatalf("session get-title should not be window-only:\n%s", resp.CapturedScript)
	}
}
```
