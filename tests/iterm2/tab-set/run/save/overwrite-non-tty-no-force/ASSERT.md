## Expected

- Non-zero exit.
- bots.json still original fixture (ids a,b; echo a / echo b; window local-bots).
- Error hints force / tty / confirm / interactive / non-interactive.

## Side Effects

- No overwrite of bots.json.

## Exit Code

- ≠ 0

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode == 0 {
		t.Fatalf("non-TTY overwrite without --force must fail; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save not accepted; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("bots.json missing: %v", readErr)
	}
	var file struct {
		WindowName string `json:"window_name"`
		Tabs       []struct {
			ID      string `json:"id"`
			Command string `json:"command"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v", jerr)
	}
	if file.WindowName != "local-bots" {
		t.Fatalf("file was modified window_name=%q", file.WindowName)
	}
	if len(file.Tabs) != 2 || file.Tabs[0].ID != "a" {
		t.Fatalf("file was modified tabs=%+v", file.Tabs)
	}
	// soft check on message
	if !strings.Contains(lower, "force") && !strings.Contains(lower, "tty") &&
		!strings.Contains(lower, "confirm") && !strings.Contains(lower, "interactive") &&
		!strings.Contains(lower, "prompt") && !strings.Contains(lower, "overwrite") {
		t.Logf("error message could mention force/tty/confirm: %s", out)
	}
}
```
