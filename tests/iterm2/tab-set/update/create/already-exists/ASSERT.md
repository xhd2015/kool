## Expected

- Non-zero exit.
- Error hints already exists / exists / duplicate (soft).
- Tab a command still `echo a` (not overwritten).

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
		t.Fatalf("create on existing id must fail; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Tabs []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v", jerr)
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("file modified tabs=%d", len(file.Tabs))
	}
	for _, tab := range file.Tabs {
		if tab.ID == "a" && tab.Command != "echo a" {
			t.Fatalf("tab a command changed to %q", tab.Command)
		}
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "exist") && !strings.Contains(lower, "already") &&
		!strings.Contains(lower, "duplicate") {
		t.Logf("error could mention already exists: %s", out)
	}
}
```
