## Expected

- Exit 0.
- File overwritten: tab b command is `echo B-changed`; tab c present; window_name new-win.
- Output mentions at least two of: unchanged, modified, added, deleted
  (diff buckets; field-level +/- preferred for command changes).

## Side Effects

- bots.json replaced under ConfigDir.

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
		t.Fatalf("overwrite-force exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save/--force not accepted; out:\n%s", out)
	}
	buckets := 0
	for _, b := range []string{"unchanged", "modified", "added", "deleted"} {
		if strings.Contains(lower, b) {
			buckets++
		}
	}
	if buckets < 2 {
		t.Fatalf("expected diff buckets (unchanged/modified/added/deleted); out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Version    int    `json:"version"`
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.WindowName != "new-win" {
		t.Fatalf("window_name want new-win got %q", file.WindowName)
	}
	byID := map[string]string{}
	for _, tab := range file.Tabs {
		byID[tab.ID] = tab.Command
	}
	if byID["a"] != "echo a" {
		t.Fatalf("tab a: %q", byID["a"])
	}
	if byID["b"] != "echo B-changed" {
		t.Fatalf("tab b want echo B-changed got %q", byID["b"])
	}
	if byID["c"] != "echo c" {
		t.Fatalf("tab c: %q", byID["c"])
	}
}
```
