## Expected Output

```
title changed:  -> fresh-name
```

## Expected

- Exit 0.
- Stdout is `title changed:  -> fresh-name\n` (two spaces around empty old, or
  equivalently empty old between `: ` and ` -> `).
- Script still written for the set operation.

## Exit Code

- 0

```go
import (
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
	// Exact form from requirement: empty old printed as-is.
	assert.Output(t, resp.Stdout, `---
version: 2
---
title changed:  -> fresh-name
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript")
	}
}
```
