## Expected

- Exit 0 (partial save on dual collapse — D2)
- Saved (or Would save path not used — live write)
- Exactly one checkpoint window (no double windows from same iterm_window_id)
- Stderr warning about dual install / collapse / no new windows / duplicate
- FileJSON has `"app"` when known (system or home canonical)

## Exit Code

- 0

```go
import (
	"encoding/json"
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
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	// Dual collapse / other path no new ids → warn (D2).
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "warning") {
		t.Fatalf("expected dual-collapse warning on stderr; stderr=%q", resp.Stderr)
	}
	// Soft match: dual / collapse / duplicate / same / install / path / no new
	hasHint := strings.Contains(se, "dual") ||
		strings.Contains(se, "collapse") ||
		strings.Contains(se, "duplicate") ||
		strings.Contains(se, "same") ||
		strings.Contains(se, "install") ||
		strings.Contains(se, "no new") ||
		strings.Contains(se, "already") ||
		strings.Contains(se, "running")
	if !hasHint {
		t.Fatalf("stderr warning should mention dual/collapse/install; stderr=%q", resp.Stderr)
	}
	if resp.FileJSON == "" {
		t.Fatal("missing FileJSON (partial save still writes)")
	}
	var raw struct {
		Windows []struct {
			App           string `json:"app"`
			ItermWindowID int64  `json:"iterm_window_id"`
		} `json:"windows"`
	}
	if err := json.Unmarshal([]byte(resp.FileJSON), &raw); err != nil {
		t.Fatalf("parse: %v\n%s", err, resp.FileJSON)
	}
	if len(raw.Windows) != 1 {
		t.Fatalf("dedupe must yield exactly 1 window, got %d", len(raw.Windows))
	}
	// App set when known.
	if !fileJSONHasApp(resp.FileJSON, fixtureAppSystem) && !fileJSONHasApp(resp.FileJSON, fixtureAppHome) {
		t.Fatalf("checkpoint should set app when known:\n%s", resp.FileJSON)
	}
	if strings.Contains(resp.FileJSON, `"app": "/Users/`) {
		t.Fatalf("home app must use ~/ form:\n%s", resp.FileJSON)
	}
}
```
