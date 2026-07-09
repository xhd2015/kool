## Expected Output

```
title changed: old-window-title -> new-window-title
```

## Expected

- Exit 0; `--window` after the title is accepted.
- Same success stdout and window-targeted script as flag-before-title.

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
	if !strings.Contains(strings.ToLower(resp.CapturedScript), "window") {
		t.Fatalf("expected window-targeted script:\n%s", resp.CapturedScript)
	}
}
```
