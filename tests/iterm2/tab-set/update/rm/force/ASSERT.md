## Expected

- Exit 0.
- File has only tab `b` (command `echo b`, cwd `/tmp`).
- Output mentions removed / tab a (soft).

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
		t.Fatalf("rm force exit=%d out:\n%s", resp.ExitCode, out)
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
			ID      string `json:"id"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
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
	if len(file.Tabs) != 1 {
		t.Fatalf("want 1 tab left, got %d: %s", len(file.Tabs), data)
	}
	if file.Tabs[0].ID != "b" {
		t.Fatalf("remaining tab want b got %q", file.Tabs[0].ID)
	}
	if file.Tabs[0].Command != "echo b" || file.Tabs[0].Cwd != "/tmp" {
		t.Fatalf("tab b changed: %+v", file.Tabs[0])
	}
}
```
