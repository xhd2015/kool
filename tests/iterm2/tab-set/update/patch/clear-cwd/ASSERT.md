## Expected

- Exit 0.
- Tab `b` cwd is empty string (or omitted); command still `echo b`.
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
		t.Fatalf("patch clear-cwd exit=%d out:\n%s", resp.ExitCode, out)
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
		Tabs []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
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
	if byID["b"].Cwd != "" {
		t.Fatalf("tab b cwd want empty got %q; file=%s", byID["b"].Cwd, data)
	}
	if byID["b"].Command != "echo b" {
		t.Fatalf("tab b command changed: %q", byID["b"].Command)
	}
	if byID["a"].Command != "echo a" {
		t.Fatalf("tab a changed: %q", byID["a"].Command)
	}
}
```
