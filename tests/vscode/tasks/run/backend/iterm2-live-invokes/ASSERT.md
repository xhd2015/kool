## Expected

- Exit 0.
- Combined output must **not** contain `live iterm2 run not configured`.
- Mock out file exists and parses as RunTabSet call JSON.
- Exactly **2** tabs; labels/commands cover Step One / Step Two markers.
- `spec.windowName` is root label `Both Steps` (or contains it).
- Set name includes `vscode-tasks` and a stable slug related to the root label.

## Side Effects

- No live iTerm; mock records the intended RunTabSet call.
- Fail closed product path: no silent local multi-run as the live backend.

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
		t.Fatalf("iterm2 live mock invoke exit=%d out:\n%s", resp.ExitCode, out)
	}
	assertNoLiveStubWarning(t, out)

	if req.ITerm2MockOut == "" {
		t.Fatal("ITerm2MockOut unset — Setup must call enableITerm2Mock")
	}
	if !mockFileExists(req.ITerm2MockOut) {
		t.Fatalf("expected mock RunTabSet JSON at %s (product KOOL_VSCODE_TASKS_ITERM2_MOCK seam)", req.ITerm2MockOut)
	}
	call := readITerm2MockCall(t, req.ITerm2MockOut)
	if n := len(call.Spec.Tabs); n != 2 {
		t.Fatalf("want 2 tabs for Both Steps composite, got %d (call=%+v)", n, call)
	}

	// Window name = root task label
	if !strings.Contains(call.Spec.WindowName, "Both Steps") && call.Spec.WindowName != "Both Steps" {
		// allow exact or soft contain
		if strings.TrimSpace(call.Spec.WindowName) == "" {
			t.Fatalf("spec.windowName empty; want root label Both Steps; call=%+v", call)
		}
		// if product uses different casing only — still require Both or Steps
		lower := strings.ToLower(call.Spec.WindowName)
		if !strings.Contains(lower, "both") {
			t.Fatalf("spec.windowName=%q want Both Steps; call=%+v", call.Spec.WindowName, call)
		}
	}

	// Set name: vscode-tasks + slug
	nameLower := strings.ToLower(call.Spec.Name)
	if !strings.Contains(nameLower, "vscode-tasks") {
		t.Fatalf("spec.name=%q should include vscode-tasks prefix; call=%+v", call.Spec.Name, call)
	}

	// Tab labels / commands cover both leaves
	blob := call.Spec.Name + " " + call.Spec.WindowName
	for _, tab := range call.Spec.Tabs {
		blob += " " + tab.ID + " " + tab.Name + " " + tab.Command + " " + tab.Cwd
	}
	hasOne := strings.Contains(blob, "Step One") || strings.Contains(blob, "KOOL_TASKS_P2_STEP_ONE") || strings.Contains(blob, "step-one") || strings.Contains(blob, "step_one")
	hasTwo := strings.Contains(blob, "Step Two") || strings.Contains(blob, "KOOL_TASKS_P2_STEP_TWO") || strings.Contains(blob, "step-two") || strings.Contains(blob, "step_two")
	if !hasOne || !hasTwo {
		t.Fatalf("mock tabs must map both leaves; blob=%q call=%+v", blob, call)
	}

	// Prefer commands match fixture echo lines when present
	cmdBlob := ""
	for _, tab := range call.Spec.Tabs {
		cmdBlob += tab.Command + "\n"
	}
	if cmdBlob != "" {
		if !strings.Contains(cmdBlob, "KOOL_TASKS_P2_STEP_ONE") || !strings.Contains(cmdBlob, "KOOL_TASKS_P2_STEP_TWO") {
			t.Logf("prefer tab commands to include fixture echo markers; commands:\n%s", cmdBlob)
		}
	}
}
```
