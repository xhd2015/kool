## Expected

- Exit 0 (or non-zero only if update unimplemented — then fail assert on file if written).
- Preferred: exit 0 with plan mentioning no_submit / tab b / dry-run.
- bots.json: tab b must **not** have no_submit true (file unchanged).

## Side Effects

- No write to bots.json.

## Exit Code

- 0 when update is implemented

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
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unknown subcommand") ||
		(strings.Contains(lower, "unrecognized") && strings.Contains(lower, "update")) {
		t.Fatalf("update not accepted (RED until implementer); out:\n%s", out)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("dry-run exit=%d out:\n%s", resp.ExitCode, out)
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
	for _, tab := range file.Tabs {
		if tab.ID == "b" && tab.NoSubmit != nil && *tab.NoSubmit {
			t.Fatalf("dry-run must not write no_submit; file=%s", data)
		}
	}
}
```
