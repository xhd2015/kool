## Expected

- Exit 0
- Stdout contains ANSI escape (`\x1b[`)
- Plan still has Would restore and resume cmds
- restored_at still null

## Errors

- None

## Exit Code

- 0

```go
import (
	"regexp"
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
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("restore --color must emit ANSI; stdout:\n%s", out)
	}
	hasGreenOrBold := strings.Contains(out, "\x1b[32m") ||
		strings.Contains(out, "\x1b[1m") ||
		strings.Contains(out, "\x1b[01m")
	if !hasGreenOrBold {
		t.Fatalf("expected green and/or bold ANSI; stdout:\n%q", out)
	}
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	// formatRestoreCommand paints argv0 green and args gray separately, so
	// strip CSI before checking the contiguous resume command text.
	plain := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(out, "")
	if !strings.Contains(plain, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
