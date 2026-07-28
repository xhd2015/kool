## Expected

- Exit 0 (no live iTerm; no config file needed).
- Plan mentions command `echo staged` and id `x` (or clear tab marker).
- Plan marks no_submit (e.g. `(no_submit)` or `no_submit`).
- Does not write scratch.json.

## Exit Code

- 0

```go
import (
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
	if resp.ExitCode != 0 {
		t.Fatalf("ad-hoc no_submit dry-run exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted; out:\n%s", out)
	}
	// Must accept no_submit as a known prop key (not "unknown key").
	if strings.Contains(lower, "unknown key") && strings.Contains(lower, "no_submit") {
		t.Fatalf("no_submit prop not accepted yet; out:\n%s", out)
	}
	if !strings.Contains(out, "echo staged") {
		t.Fatalf("plan missing command; out:\n%s", out)
	}
	if !strings.Contains(lower, "no_submit") {
		t.Fatalf("plan must mark no_submit; out:\n%s", out)
	}
	if strings.Contains(lower, "not found") && strings.Contains(lower, "scratch") {
		t.Fatalf("ad-hoc must not require config file; out:\n%s", out)
	}
	if _, statErr := os.Stat(configPath(req.ConfigDir, "scratch")); statErr == nil {
		t.Fatalf("dry-run must not write scratch.json")
	}
}
```
