## Expected

- Non-zero exit.
- File still has tab `x` with command `echo solo`.
- Error mentions last / empty / cannot remove (soft).

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
		t.Fatalf("rm last tab must fail; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "solo"))
	if readErr != nil {
		t.Fatalf("read solo.json: %v", readErr)
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
	if len(file.Tabs) != 1 || file.Tabs[0].ID != "x" || file.Tabs[0].Command != "echo solo" {
		t.Fatalf("file was modified: %s", data)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "last") && !strings.Contains(lower, "empty") &&
		!strings.Contains(lower, "cannot") && !strings.Contains(lower, "at least") &&
		!strings.Contains(lower, "remain") {
		t.Logf("error could mention last/empty tabs: %s", out)
	}
}
```
