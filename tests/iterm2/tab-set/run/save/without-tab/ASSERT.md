## Expected

- Non-zero exit.
- Message mentions save and/or tab / --tab required.
- bots.json unchanged (still original).

## Exit Code

- ≠ 0

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
	if resp.ExitCode == 0 {
		t.Fatalf("--save without --tab must error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		// --save not wired yet still counts as RED for this leaf
		t.Fatalf("--save not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "save") && !strings.Contains(lower, "tab") {
		t.Fatalf("expected save/--tab error message; out:\n%s", out)
	}
	// ensure file still exists (not deleted)
	if _, e := os.Stat(configPath(req.ConfigDir, "bots")); e != nil {
		t.Fatalf("bots.json should remain: %v", e)
	}
}
```
