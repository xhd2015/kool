## Expected

- Exit 0
- Would restore + grok resume cmd
- Restore does not use app as placement target (no app-install path as action)
- restored_at still null
- Seed file still readable (not required to strip app)

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
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	// Restore ignores app: must not surface app path as a placement/target action.
	// (space meta may still appear; app line is save-plan only.)
	for _, line := range strings.Split(out, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "app ") || strings.HasPrefix(trim, "app\t") {
			t.Fatalf("restore dry-run must not show app meta/placement line:\n%s", out)
		}
	}
	// Must not error because of unknown app field.
	if strings.Contains(strings.ToLower(resp.Stderr), "unknown") && strings.Contains(strings.ToLower(resp.Stderr), "app") {
		t.Fatalf("restore must ignore app field; stderr=%q", resp.Stderr)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
