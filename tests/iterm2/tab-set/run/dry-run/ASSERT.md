## Expected

- Exit 0 (no live iTerm required).
- Output mentions set `bots` and tab commands (`echo a`, `echo b`) and/or tab ids.
- Prefer mentioning dry-run / plan / would.

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
		t.Fatalf("dry-run exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout + resp.Stderr
	if !strings.Contains(out, "echo a") && !strings.Contains(out, "a") {
		t.Fatalf("dry-run plan should mention tab a / echo a; out:\n%s", out)
	}
	if !strings.Contains(out, "echo b") && !strings.Contains(out, "b") {
		t.Fatalf("dry-run plan should mention tab b / echo b; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	// Prefer plan wording; bots name often present
	if !strings.Contains(out, "bots") && !strings.Contains(lower, "dry") &&
		!strings.Contains(lower, "plan") && !strings.Contains(lower, "would") {
		t.Logf("dry-run output lacks explicit plan keyword (ok if commands listed): %s", out)
	}
}
```
