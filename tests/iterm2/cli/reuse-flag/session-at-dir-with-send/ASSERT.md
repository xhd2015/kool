## Expected

- Exit 0; script may include `write text "grok"` only in the miss (else) branch.
- Match branch must not run follow-ups when reusing an existing session at `targetDir`.

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
	if !strings.Contains(s, writeTextCmd("grok")) {
		t.Fatalf("script should still emit grok for miss branch: %q", s)
	}
	matchBranchMustNotContain(t, s, writeTextCmd("grok"))
	matchBranchMustNotContain(t, s, writeTextCd())
}
```