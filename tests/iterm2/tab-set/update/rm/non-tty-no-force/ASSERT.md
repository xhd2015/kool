## Expected

- Non-zero exit.
- bots.json still has both tabs a and b.
- Error hints force / tty / confirm / interactive (soft).

## Side Effects

- No write (file unchanged).

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
		t.Fatalf("non-TTY rm without --force must fail; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Tabs []struct {
			ID string `json:"id"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v", jerr)
	}
	if len(file.Tabs) != 2 {
		t.Fatalf("file was modified tabs=%d; out:\n%s", len(file.Tabs), out)
	}
	ids := map[string]bool{}
	for _, tab := range file.Tabs {
		ids[tab.ID] = true
	}
	if !ids["a"] || !ids["b"] {
		t.Fatalf("expected tabs a,b unchanged: %v", ids)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "force") && !strings.Contains(lower, "tty") &&
		!strings.Contains(lower, "confirm") && !strings.Contains(lower, "interactive") &&
		!strings.Contains(lower, "prompt") {
		t.Logf("error could mention force/tty/confirm: %s", out)
	}
}
```
