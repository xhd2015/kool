## Expected

- Non-zero exit.
- Error mentions command / --command / required (soft).
- bots.json still 2 tabs (a, b).

## Side Effects

- No new tab written.

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
		t.Fatalf("create without --command must fail; out:\n%s", out)
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
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "command") && !strings.Contains(lower, "create") {
		t.Logf("error could mention command/create: %s", out)
	}
}
```
