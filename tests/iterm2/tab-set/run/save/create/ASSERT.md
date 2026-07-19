## Expected

- Exit 0.
- File `ConfigDir/mysave.json` exists with version 1, window_name win-save,
  tabs x/y with commands echo x / echo y; y may have cwd `/tmp`.
- No iTerm run required (success without live iTerm proves save-only path).

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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("save create exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save not accepted; out:\n%s", out)
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
			ID      string `json:"id"`
			Name    string `json:"name"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("invalid JSON written: %v\n%s", jerr, data)
	}
	if file.Version != 1 {
		t.Fatalf("version want 1 got %d", file.Version)
	}
	if file.WindowName != "win-save" {
		t.Fatalf("window_name want win-save got %q", file.WindowName)
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("want 2 tabs, got %d: %s", len(file.Tabs), data)
	}
	byID := map[string]struct {
		Command string
		Cwd     string
	}{}
	for _, tab := range file.Tabs {
		byID[tab.ID] = struct {
			Command string
			Cwd     string
		}{tab.Command, tab.Cwd}
	}
	if byID["x"].Command != "echo x" {
		t.Fatalf("tab x command: %+v file=%s", byID["x"], data)
	}
	if byID["y"].Command != "echo y" {
		t.Fatalf("tab y command: %+v file=%s", byID["y"], data)
	}
	if byID["y"].Cwd != "/tmp" {
		t.Fatalf("tab y cwd want /tmp got %q", byID["y"].Cwd)
	}
}
```
