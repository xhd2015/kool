## Expected

- Exit 0.
- `window_name` is `new-bots-win`.
- Tabs a/b ids, names, commands, cwd unchanged.

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
		t.Fatalf("window-name-only exit=%d out:\n%s", resp.ExitCode, out)
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
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.WindowName != "new-bots-win" {
		t.Fatalf("window_name want new-bots-win got %q", file.WindowName)
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("tabs changed count=%d", len(file.Tabs))
	}
	byID := map[string]string{}
	byCwd := map[string]string{}
	for _, tab := range file.Tabs {
		byID[tab.ID] = tab.Command
		byCwd[tab.ID] = tab.Cwd
	}
	if byID["a"] != "echo a" || byID["b"] != "echo b" {
		t.Fatalf("tab commands changed: %v", byID)
	}
	if byCwd["b"] != "/tmp" {
		t.Fatalf("tab b cwd changed: %q", byCwd["b"])
	}
}
```
