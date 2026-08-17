## Expected

- Exit 0; else branch includes `write text "grok"` after cd.
- Match branch still has no follow-up commands.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	s := resp.CapturedScript
	if s == "" {
		t.Fatal("expected captured script")
	}
	if !scriptHasReusePathScan(s) {
		t.Fatalf("missing path scan: %q", s)
	}
	elseBranchMustContain(t, s, writeTextCmd("grok"))
	matchBranchMustNotContain(t, s, writeTextCmd("grok"))
	if strings.Contains(s, "current session of current tab of current window") {
		t.Fatal("reuse -r must not use legacy current-session cd target")
	}
}
```