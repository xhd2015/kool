## Expected Output

```
title changed: prev -> say "hi"\path
```

## Expected

- Exit 0; success line shows the new title as the user supplied it.
- Captured script escapes `"` as `\"` and `\` as `\\` for AppleScript string
  literals (same rules as path/command escaping in `shell/iterm2`).

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
title changed: prev -> say "hi"\path
`)
	if !resp.ScriptWritten || resp.CapturedScript == "" {
		t.Fatal("expected captured AppleScript")
	}
	// Escaped forms must appear in the script body.
	if !strings.Contains(resp.CapturedScript, `\"`) {
		t.Fatalf("expected escaped double-quote in script:\n%s", resp.CapturedScript)
	}
	if !strings.Contains(resp.CapturedScript, `\\`) {
		t.Fatalf("expected escaped backslash in script:\n%s", resp.CapturedScript)
	}
	// Unescaped raw quote after a set-to open quote is unsafe; require no bare
	// `say "hi"` without backslash-escape.
	if strings.Contains(resp.CapturedScript, `say "hi"`) {
		t.Fatalf("raw unescaped quotes in script:\n%s", resp.CapturedScript)
	}
}
```
