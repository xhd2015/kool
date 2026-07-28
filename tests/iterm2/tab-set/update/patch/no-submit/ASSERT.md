## Expected

- Exit 0.
- File `bots.json`: tab `b` has `"no_submit": true`; tab `a` omits or false.
- Tab a/b commands and window_name unchanged.
- Stdout/stderr mentions update / tab b / no_submit (soft).

## Side Effects

- bots.json rewritten under ConfigDir.

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
		t.Fatalf("patch no-submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unknown subcommand") || strings.Contains(lower, "unrecognized") {
		t.Fatalf("update not accepted; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Version    int    `json:"version"`
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID       string `json:"id"`
			Command  string `json:"command"`
			Cwd      string `json:"cwd"`
			NoSubmit *bool  `json:"no_submit"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.Version != 1 {
		t.Fatalf("version want 1 got %d", file.Version)
	}
	if file.WindowName != "local-bots" {
		t.Fatalf("window_name changed: %q", file.WindowName)
	}
	byID := map[string]struct {
		Command  string
		Cwd      string
		NoSubmit *bool
	}{}
	for i := range file.Tabs {
		tab := &file.Tabs[i]
		byID[tab.ID] = struct {
			Command  string
			Cwd      string
			NoSubmit *bool
		}{tab.Command, tab.Cwd, tab.NoSubmit}
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("want 2 tabs, got %d: %s", len(file.Tabs), data)
	}
	if byID["a"].Command != "echo a" {
		t.Fatalf("tab a command changed: %q", byID["a"].Command)
	}
	if byID["a"].NoSubmit != nil && *byID["a"].NoSubmit {
		t.Fatalf("tab a must not be no_submit; file=%s", data)
	}
	if byID["b"].Command != "echo b" {
		t.Fatalf("tab b command changed: %q", byID["b"].Command)
	}
	if byID["b"].Cwd != "/tmp" {
		t.Fatalf("tab b cwd changed: %q", byID["b"].Cwd)
	}
	if byID["b"].NoSubmit == nil || !*byID["b"].NoSubmit {
		t.Fatalf("tab b want no_submit=true; file=%s", data)
	}
}
```
