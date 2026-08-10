## Expected

- Exit 0
- Saved
- One window kept (space 0); filter.spaces includes 0
- Kept window has system app `/Applications/iTerm.app`
- Stderr skip warning for --spaces
- Home app window dropped (no `~/Applications/iTerm.app` required on remaining)

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
	if !strings.Contains(resp.Stderr, "skipped") || !strings.Contains(resp.Stderr, "--spaces") {
		t.Fatalf("expected --spaces skip warning; stderr=%q", resp.Stderr)
	}
	if resp.FileJSON == "" {
		t.Fatal("missing FileJSON")
	}
	var raw struct {
		Filter *struct {
			Spaces []int `json:"spaces"`
		} `json:"filter"`
		Summary struct {
			Windows  int `json:"windows"`
			Sessions int `json:"sessions"`
		} `json:"summary"`
		Windows []struct {
			App   string `json:"app"`
			Space int    `json:"space"`
		} `json:"windows"`
	}
	if err := json.Unmarshal([]byte(resp.FileJSON), &raw); err != nil {
		t.Fatalf("parse: %v\n%s", err, resp.FileJSON)
	}
	if raw.Filter == nil || len(raw.Filter.Spaces) != 1 || raw.Filter.Spaces[0] != 0 {
		t.Fatalf("filter=%+v", raw.Filter)
	}
	if len(raw.Windows) != 1 || raw.Summary.Windows != 1 {
		t.Fatalf("want 1 kept window; windows=%d summary=%+v", len(raw.Windows), raw.Summary)
	}
	if raw.Windows[0].Space != 0 {
		t.Fatalf("kept window space=%d want 0", raw.Windows[0].Space)
	}
	if !fileJSONHasApp(resp.FileJSON, fixtureAppSystem) {
		t.Fatalf("kept window must carry system app:\n%s", resp.FileJSON)
	}
	// Dropped home window should not leave home app on the sole remaining window.
	if fileJSONHasApp(resp.FileJSON, fixtureAppHome) {
		t.Fatalf("kept-only checkpoint should not include home app:\n%s", resp.FileJSON)
	}
}
```
