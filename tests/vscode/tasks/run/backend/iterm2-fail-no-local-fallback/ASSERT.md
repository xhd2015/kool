## Expected

- Non-zero exit (mock forces RunTabSet / backend error).
- Combined output must **not** contain **both** `KOOL_TASKS_P2_STEP_ONE` and
  `KOOL_TASKS_P2_STEP_TWO` (would imply silent local multi-run fallback).
- Prefer: no `live iterm2 run not configured` stub path (error instead of local).

## Errors

- Fail closed: iterm2 backend error surfaces as failure, not local success.

## Exit Code

- ≠ 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode == 0 {
		t.Fatalf("mock err must fail closed (exit≠0); out:\n%s", out)
	}

	hasOne := strings.Contains(out, "KOOL_TASKS_P2_STEP_ONE")
	hasTwo := strings.Contains(out, "KOOL_TASKS_P2_STEP_TWO")
	if hasOne && hasTwo {
		t.Fatalf("silent local multi-leaf fallback detected (both step markers); out:\n%s", out)
	}

	// Current product stub prints warning then local-runs — still RED until fixed.
	lower := strings.ToLower(out)
	if strings.Contains(lower, "live iterm2 run not configured") && hasOne {
		t.Fatalf("stub local fallback after live-not-configured; want fail closed; out:\n%s", out)
	}
}
```
