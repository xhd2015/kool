## Expected

- Exit 0
- Would restore + grok resume cmd
- Default mode (no `--same-app`):
  - global **`restore target`** prefers home when both installs exist
  - **`recorded app`** system shown because it differs from target
  - no same-app create **`app  …`** line (recorded app is not used for create)
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
	// Must not error because of unknown app field.
	if strings.Contains(strings.ToLower(resp.Stderr), "unknown") &&
		strings.Contains(strings.ToLower(resp.Stderr), "app") {
		t.Fatalf("restore must accept app field; stderr=%q", resp.Stderr)
	}

	// New contract: default prefers home create target; honesty meta for recorded.
	// Helpers live under restore/app-target; re-implement lightly here so this
	// sibling leaf does not depend on app-target SETUP inheritance.
	if !lineHas(out, "restore target", fixtureAppHome) {
		t.Fatalf("default both-installs: expected restore target %s:\n%s", fixtureAppHome, out)
	}
	if lineHas(out, "restore target", fixtureAppSystem) {
		t.Fatalf("default both-installs must not restore-target system:\n%s", out)
	}
	if !lineHas(out, "recorded app", fixtureAppSystem) {
		t.Fatalf("expected recorded app %s when it differs from target:\n%s", fixtureAppSystem, out)
	}
	// No same-app create line (R4 — recorded app not used for create).
	for _, line := range strings.Split(out, "\n") {
		trim := strings.TrimSpace(line)
		// rough ANSI strip not required for monochrome default
		if strings.HasPrefix(trim, "recorded app") || strings.Contains(trim, "restore target") {
			continue
		}
		if strings.HasPrefix(trim, "app ") || strings.HasPrefix(trim, "app\t") {
			t.Fatalf("default mode must not show same-app create app line:\n%s", out)
		}
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}

func lineHas(out, label, app string) bool {
	for _, line := range strings.Split(out, "\n") {
		trim := strings.TrimSpace(line)
		idx := strings.Index(trim, label)
		if idx < 0 {
			continue
		}
		// Exact path token after label so "~/Applications/…" ≠ "/Applications/…".
		rest := strings.TrimSpace(trim[idx+len(label):])
		fields := strings.Fields(rest)
		if len(fields) > 0 && fields[0] == app {
			return true
		}
	}
	return false
}
```
