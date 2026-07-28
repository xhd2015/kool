## Expected

- Exit 0.
- File: tab `b` has no_submit omitted or false; command/cwd unchanged.
- Tab `a` unchanged.

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
		t.Fatalf("patch submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unknown subcommand") || strings.Contains(lower, "unrecognized") {
		t.Fatalf("update not accepted; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "staged"))
	if readErr != nil {
		t.Fatalf("read staged.json: %v", readErr)
	}
	var file struct {
		Tabs []struct {
			ID       string `json:"id"`
			Command  string `json:"command"`
			Cwd      string `json:"cwd"`
			NoSubmit *bool  `json:"no_submit"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
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
	if byID["b"].Command != "grok --resume" {
		t.Fatalf("tab b command: %q", byID["b"].Command)
	}
	if byID["b"].Cwd != "/tmp" {
		t.Fatalf("tab b cwd: %q", byID["b"].Cwd)
	}
	if byID["b"].NoSubmit != nil && *byID["b"].NoSubmit {
		t.Fatalf("tab b want no_submit cleared; file=%s", data)
	}
	if byID["a"].Command != "echo a" {
		t.Fatalf("tab a changed: %q", byID["a"].Command)
	}
}
```
