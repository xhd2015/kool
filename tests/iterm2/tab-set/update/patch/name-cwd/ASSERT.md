## Expected

- Exit 0.
- Tab `a`: name `Alpha`, cwd `/var/tmp`, command still `echo a`.
- Tab `b` unchanged.

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
		t.Fatalf("patch name-cwd exit=%d out:\n%s", resp.ExitCode, out)
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
			Name    string `json:"name"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
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
	if byID["a"].Name != "Alpha" {
		t.Fatalf("tab a name want Alpha got %q", byID["a"].Name)
	}
	if byID["a"].Cwd != "/var/tmp" {
		t.Fatalf("tab a cwd want /var/tmp got %q", byID["a"].Cwd)
	}
	if byID["a"].Command != "echo a" {
		t.Fatalf("tab a command changed: %q", byID["a"].Command)
	}
	if byID["b"].Name != "b" || byID["b"].Command != "echo b" || byID["b"].Cwd != "/tmp" {
		t.Fatalf("tab b changed: %+v", byID["b"])
	}
}
```
