## Expected

- Exit 0 (offline; no live iTerm).
- Output mentions both leaf labels and/or their echo commands (`Step One`,
  `Step Two`, `KOOL_TASKS_P2_STEP_ONE`, `KOOL_TASKS_P2_STEP_TWO`).
- Prefer wording for dry-run / plan / tab (tab-set style).
- Mock out file must **not** exist (dry-run never calls RunTabSet).

## Side Effects

- No live iTerm; no RunTabSet; mock JSON absent.

## Exit Code

- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("iterm2 dry-run tabs exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--backend=iterm2 not accepted; out:\n%s", out)
	}
	// Both leaves should appear in the tab/plan output.
	hasOne := strings.Contains(out, "Step One") || strings.Contains(out, "KOOL_TASKS_P2_STEP_ONE") || strings.Contains(out, "echo KOOL_TASKS_P2_STEP_ONE")
	hasTwo := strings.Contains(out, "Step Two") || strings.Contains(out, "KOOL_TASKS_P2_STEP_TWO") || strings.Contains(out, "echo KOOL_TASKS_P2_STEP_TWO")
	if !hasOne || !hasTwo {
		t.Fatalf("iterm2 dry-run tab plan should mention both leaf steps; out:\n%s", out)
	}
	// Prefer dry-run / tab language when present (soft preference via log if missing).
	if !strings.Contains(lower, "tab") && !strings.Contains(lower, "dry") && !strings.Contains(lower, "plan") {
		t.Logf("prefer dry-run/tab/plan wording in iterm2 dry-run output; out:\n%s", out)
	}
	// Dry-run must not invoke RunTabSet mock writer.
	if req.ITerm2MockOut != "" && mockFileExists(req.ITerm2MockOut) {
		t.Fatalf("dry-run must not write RunTabSet mock JSON at %s", req.ITerm2MockOut)
	}
}
```
