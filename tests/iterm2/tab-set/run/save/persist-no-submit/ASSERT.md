## Expected

- Exit 0.
- File `ConfigDir/mysave.json` exists with version 1.
- Tab y has `"no_submit": true` (JSON true).
- Tab x either omits `no_submit` or has false.
- No live iTerm required.

## Side Effects

- Creates mysave.json under KOOL_ITERM2_TAB_SET_DIR.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("save persist no_submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save not accepted; out:\n%s", out)
	}
	if strings.Contains(lower, "unknown key") && strings.Contains(lower, "no_submit") {
		t.Fatalf("no_submit prop not accepted on save path; out:\n%s", out)
	}
	path := configPath(req.ConfigDir, "mysave")
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected mysave.json written: %v; out:\n%s", readErr, out)
	}
	var file struct {
		Version    int    `json:"version"`
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID       string `json:"id"`
			Command  string `json:"command"`
			NoSubmit *bool  `json:"no_submit"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("invalid JSON written: %v\n%s", jerr, data)
	}
	if file.Version != 1 {
		t.Fatalf("version want 1 got %d", file.Version)
	}
	byID := map[string]*bool{}
	byCmd := map[string]string{}
	for i := range file.Tabs {
		tab := &file.Tabs[i]
		byID[tab.ID] = tab.NoSubmit
		byCmd[tab.ID] = tab.Command
	}
	if byCmd["x"] != "echo x" {
		t.Fatalf("tab x command: %q file=%s", byCmd["x"], data)
	}
	if byCmd["y"] != "echo y" {
		t.Fatalf("tab y command: %q file=%s", byCmd["y"], data)
	}
	// Tab y must persist no_submit true.
	if byID["y"] == nil || !*byID["y"] {
		t.Fatalf("tab y want no_submit=true; file=%s", data)
	}
	// Tab x: omit or false.
	if byID["x"] != nil && *byID["x"] {
		t.Fatalf("tab x must not have no_submit=true; file=%s", data)
	}
}
```
