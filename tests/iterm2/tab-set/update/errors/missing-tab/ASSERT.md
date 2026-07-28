## Expected

- Non-zero exit.
- Error mentions missing tab / not found; prefer hint `--create`.
- bots.json still 2 tabs without no_submit true.

## Exit Code

- ≠ 0

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

	// Classic TDD: fail while update is not a known subcommand (do not
	// treat "unknown subcommand" as the scenario-specific error).
	if strings.Contains(strings.ToLower(out), "unknown subcommand") {
		t.Fatalf("update not implemented yet (RED until implementer); out:\n%s", out)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("missing tab must fail; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "missing") && !strings.Contains(lower, "not found") &&
		!strings.Contains(lower, "unknown") {
		t.Fatalf("expected missing-tab error; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Tabs []struct {
			ID       string `json:"id"`
			NoSubmit *bool  `json:"no_submit"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v", jerr)
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("file modified tabs=%d", len(file.Tabs))
	}
	for _, tab := range file.Tabs {
		if tab.NoSubmit != nil && *tab.NoSubmit {
			t.Fatalf("file was patched: %s", data)
		}
	}
}
```
