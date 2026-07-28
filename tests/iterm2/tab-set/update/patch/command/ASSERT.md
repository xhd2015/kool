## Expected

- Exit 0.
- Tab `a` command is `echo A-new`; name still `a`.
- Tab `b` fully unchanged.

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
		t.Fatalf("patch command exit=%d out:\n%s", resp.ExitCode, out)
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
			Name    string `json:"name"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.WindowName != "local-bots" {
		t.Fatalf("window_name changed: %q", file.WindowName)
	}
	byID := map[string]struct {
		Name    string
		Command string
		Cwd     string
	}{}
	for _, tab := range file.Tabs {
		byID[tab.ID] = struct {
			Name    string
			Command string
			Cwd     string
		}{tab.Name, tab.Command, tab.Cwd}
	}
	if byID["a"].Command != "echo A-new" {
		t.Fatalf("tab a command want echo A-new got %q", byID["a"].Command)
	}
	if byID["a"].Name != "a" {
		t.Fatalf("tab a name changed: %q", byID["a"].Name)
	}
	if byID["b"].Command != "echo b" || byID["b"].Cwd != "/tmp" {
		t.Fatalf("tab b changed: %+v", byID["b"])
	}
}
```
