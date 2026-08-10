## Expected

- Exit 0
- Stdout does not start with a blank line
- Stdout contains app meta line with FixtureApp (e.g. `app  /Applications/iTerm.app`)
- Would save present
- app-plan.json not created

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"regexp"
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
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if strings.HasPrefix(out, "\n") {
		t.Fatalf("dry-run must not start with a leading empty line; stdout:\n%q", out)
	}
	if !strings.Contains(out, "Would save") {
		t.Fatalf("missing Would save:\n%s", out)
	}
	want := req.FixtureApp
	if want == "" {
		want = fixtureAppSystem
	}
	// Gray meta line shape: "app  <canonical>" (flexible whitespace).
	re := regexp.MustCompile(`(?m)^\s*app\s+` + regexp.QuoteMeta(want) + `\s*$`)
	if !re.MatchString(out) {
		// Also accept inline form without strict line anchors if painted.
		if !strings.Contains(out, "app") || !strings.Contains(out, want) {
			t.Fatalf("dry-run plan missing app meta for %q:\n%s", want, out)
		}
		// Require both tokens on same line when loose match.
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "app") && strings.Contains(line, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("dry-run plan missing app meta line for %q:\n%s", want, out)
		}
	}
	p := filepath.Join(req.WorkingDir, "app-plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
