## Expected

- Non-zero exit.
- Error mentions nothing to update / no changes / require (soft).
- bots.json unchanged (window_name still local-bots; 2 tabs).

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
		t.Fatalf("nothing-to-update must fail; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID string `json:"id"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v", jerr)
	}
	if file.WindowName != "local-bots" || len(file.Tabs) != 2 {
		t.Fatalf("file was modified: %s", data)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "nothing") && !strings.Contains(lower, "no change") &&
		!strings.Contains(lower, "no action") && !strings.Contains(lower, "required") &&
		!strings.Contains(lower, "specify") && !strings.Contains(lower, "empty") {
		t.Logf("error could mention nothing to update: %s", out)
	}
}
```
