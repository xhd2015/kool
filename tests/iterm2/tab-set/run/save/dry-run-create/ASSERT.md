## Expected

- Exit 0.
- Output mentions create / would create / dry-run (save plan).
- File `newsset.json` must **not** exist after run.

## Side Effects

- No disk write under ConfigDir for newsset.

## Exit Code

- 0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("save dry-run exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "create") && !strings.Contains(lower, "would") &&
		!strings.Contains(lower, "dry") && !strings.Contains(lower, "plan") {
		t.Fatalf("expected create/dry-run plan wording; out:\n%s", out)
	}
	if _, statErr := os.Stat(configPath(req.ConfigDir, "newsset")); statErr == nil {
		t.Fatalf("--save --dry-run must not write newsset.json")
	}
	// Prefer command visible in plan
	if !strings.Contains(out, "echo only") && !strings.Contains(out, "tab-1") {
		t.Logf("plan may still omit command text: %s", out)
	}
}
```
