## Expected

- Exit 0 (no live iTerm; no config file needed).
- Combined output mentions both commands (`echo alpha`, `echo beta`).
- Prefer dry-run / plan / would wording and set name `scratch`.

## Side Effects

- No `scratch.json` created under ConfigDir (ad-hoc dry-run does not save).

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
		t.Fatalf("ad-hoc dry-run exit=%d out:\n%s", resp.ExitCode, out)
	}
	// Must not fall through as unknown flag / open-dir.
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted yet (or unknown flag); out:\n%s", out)
	}
	for _, want := range []string{"echo alpha", "echo beta"} {
		if !strings.Contains(out, want) {
			t.Fatalf("plan missing %q; out:\n%s", want, out)
		}
	}
	// Must not require missing config in ad-hoc mode.
	if strings.Contains(lower, "not found") && strings.Contains(lower, "scratch") {
		t.Fatalf("ad-hoc must not require config file; out:\n%s", out)
	}
	if _, statErr := os.Stat(configPath(req.ConfigDir, "scratch")); statErr == nil {
		t.Fatalf("dry-run must not write scratch.json")
	}
}
```
