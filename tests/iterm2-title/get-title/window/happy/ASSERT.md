## Expected Output

```
Project Window
```

## Expected

- Exit 0; stdout title + trailing newline.
- Script includes UUID and references window name.

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
Project Window
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript for get-title --window")
	}
	uuid := sessionUUID(req)
	if !strings.Contains(resp.CapturedScript, uuid) {
		t.Fatalf("script should locate session UUID %q; got:\n%s", uuid, resp.CapturedScript)
	}
	lower := strings.ToLower(resp.CapturedScript)
	if !strings.Contains(lower, "window") {
		t.Fatalf("--window get should reference window:\n%s", resp.CapturedScript)
	}
}
```
