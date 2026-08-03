## Expected

- Exit 0
- Stdout contains `space` + `Desktop` label in the form `space N (Desktop N+1)`
- Stdout does **not** mention `iterm_window_id`
- space-plan.json not created

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
	// space N (Desktop N+1) — N is a non-negative integer
	re := regexp.MustCompile(`space\s+\d+\s+\(Desktop\s+\d+\)`)
	if !re.MatchString(out) {
		t.Fatalf("dry-run plan missing space N (Desktop N+1) label:\n%s", out)
	}
	if strings.Contains(out, "iterm_window_id") {
		t.Fatalf("dry-run must not show iterm_window_id:\n%s", out)
	}
	if !strings.Contains(out, "Would save") {
		t.Fatalf("missing Would save:\n%s", out)
	}
	p := filepath.Join(req.WorkingDir, "space-plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
