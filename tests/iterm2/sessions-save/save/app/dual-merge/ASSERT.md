## Expected

- Exit 0
- Saved
- FileJSON has both `"app": "/Applications/iTerm.app"` and `"app": "~/Applications/iTerm.app"`
- No `"app": "/Users/…"`
- Two windows; no duplicate `iterm_window_id`; summary windows=2, sessions=2
- source_index renumbered globally 1…N (present and unique)

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
	if resp.FileJSON == "" {
		t.Fatal("missing FileJSON")
	}
	if !fileJSONHasApp(resp.FileJSON, fixtureAppSystem) {
		t.Fatalf("missing system app in FileJSON:\n%s", resp.FileJSON)
	}
	if !fileJSONHasApp(resp.FileJSON, fixtureAppHome) {
		t.Fatalf("missing home app (~/ form) in FileJSON:\n%s", resp.FileJSON)
	}
	if strings.Contains(resp.FileJSON, `"app": "/Users/`) {
		t.Fatalf("home app must use ~/ form, not /Users/…:\n%s", resp.FileJSON)
	}

	// Parse windows for counts + id uniqueness without requiring SaveWindow.App.
	var raw struct {
		Summary struct {
			Windows  int            `json:"windows"`
			Sessions int            `json:"sessions"`
			ByKind   map[string]int `json:"by_kind"`
		} `json:"summary"`
		Windows []struct {
			SourceIndex   int    `json:"source_index"`
			App           string `json:"app"`
			ItermWindowID int64  `json:"iterm_window_id"`
		} `json:"windows"`
	}
	if err := json.Unmarshal([]byte(resp.FileJSON), &raw); err != nil {
		t.Fatalf("FileJSON parse: %v\n%s", err, resp.FileJSON)
	}
	if len(raw.Windows) != 2 {
		t.Fatalf("want 2 windows, got %d", len(raw.Windows))
	}
	if raw.Summary.Windows != 2 || raw.Summary.Sessions != 2 {
		t.Fatalf("summary windows=%d sessions=%d by_kind=%v", raw.Summary.Windows, raw.Summary.Sessions, raw.Summary.ByKind)
	}
	seenID := map[int64]bool{}
	seenIdx := map[int]bool{}
	apps := map[string]bool{}
	for _, w := range raw.Windows {
		if w.ItermWindowID != 0 {
			if seenID[w.ItermWindowID] {
				t.Fatalf("duplicate iterm_window_id %d", w.ItermWindowID)
			}
			seenID[w.ItermWindowID] = true
		}
		if w.SourceIndex < 1 || seenIdx[w.SourceIndex] {
			t.Fatalf("bad/duplicate source_index %d", w.SourceIndex)
		}
		seenIdx[w.SourceIndex] = true
		if w.App != "" {
			apps[w.App] = true
		}
	}
	if !apps[fixtureAppSystem] || !apps[fixtureAppHome] {
		t.Fatalf("window app tags=%v want both canonical apps", apps)
	}
}
```
