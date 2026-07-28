## Expected

- Exit 0.
- File has 3 tabs; id `c` with command `echo c`.
- Tabs a/b unchanged; version 1.

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
		t.Fatalf("create with-command exit=%d out:\n%s", resp.ExitCode, out)
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
		Version int `json:"version"`
		Tabs    []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.Version != 1 {
		t.Fatalf("version want 1 got %d", file.Version)
	}
	if len(file.Tabs) != 3 {
		t.Fatalf("want 3 tabs, got %d: %s", len(file.Tabs), data)
	}
	byID := map[string]string{}
	for _, tab := range file.Tabs {
		byID[tab.ID] = tab.Command
	}
	if byID["a"] != "echo a" || byID["b"] != "echo b" {
		t.Fatalf("existing tabs changed: %v", byID)
	}
	if byID["c"] != "echo c" {
		t.Fatalf("tab c want echo c got %q; file=%s", byID["c"], data)
	}
}
```
