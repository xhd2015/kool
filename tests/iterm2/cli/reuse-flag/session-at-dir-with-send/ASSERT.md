## Expected

- Exit 0; script may include `write text "grok"` only in the miss (else) branch.
- Match branch must not run follow-ups when reusing an existing session at `targetDir`.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	if !strings.Contains(s, `write text "grok"`) {
		t.Fatalf("script should still emit grok for miss branch: %q", s)
	}
	matchBranchMustNotContain(t, s, `write text "grok"`)
	matchBranchMustNotContain(t, s, `write text ("cd " & quoted form of targetDir)`)
}
```