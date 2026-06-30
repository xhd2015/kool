## Expected

- Exit 0; else (miss) branch creates a new window, runs `cd`, and **registers**
  `targetDir` on the new session via `user.koolTargetDir` so a back-to-back second
  invocation can match before shell `path` catches up.
- Match branch must not register (focus only).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

const koolTargetVar = `variable named "user.koolTargetDir"`

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
	elseBranch := reuseScriptElseBranch(s)
	if elseBranch == "" {
		t.Fatalf("missing else branch: %q", s)
	}
	if !strings.Contains(elseBranch, koolTargetVar) {
		t.Fatalf("else branch must register %s for immediate reuse; branch=%q", koolTargetVar, elseBranch)
	}
	if !strings.Contains(elseBranch, "to targetDir") {
		t.Fatalf("else branch must assign targetDir to %s; branch=%q", koolTargetVar, elseBranch)
	}
	match := reuseScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if strings.Contains(match, koolTargetVar) {
		t.Fatalf("match branch must not register %s: %q", koolTargetVar, match)
	}
}
```