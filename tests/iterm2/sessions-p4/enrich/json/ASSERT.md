## Expected

- Exit 0.
- Stdout is one JSON object with `windows` array.
- Some session under windows/tabs has `agent` object with:
  - `session_id` = `019fabcdef-1234-5678-9abc-def012345678`
  - `kind` = `grok` (when present)
- Prefer also `tree` array non-empty (pids 100/101 from busyGrokResolve), but
  **hard requirement** is `agent.session_id`.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if strings.Contains(out, "\x1b[") {
		t.Fatal("ANSI in JSON output")
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("stdout is not a single JSON document: %v\n%s", err, out)
	}
	// Walk sessions for agent.session_id.
	found := false
	var foundKind string
	var treeLen int
	wins, _ := doc["windows"].([]any)
	for _, w := range wins {
		wm, _ := w.(map[string]any)
		tabs, _ := wm["tabs"].([]any)
		for _, tab := range tabs {
			tm, _ := tab.(map[string]any)
			sessions, _ := tm["sessions"].([]any)
			for _, s := range sessions {
				sm, _ := s.(map[string]any)
				agent, ok := sm["agent"].(map[string]any)
				if !ok || agent == nil {
					continue
				}
				sid, _ := agent["session_id"].(string)
				if sid == fixtureGrokSessionID {
					found = true
					foundKind, _ = agent["kind"].(string)
					if tr, ok := agent["tree"].([]any); ok {
						treeLen = len(tr)
					}
				}
			}
		}
	}
	if !found {
		// Also accept flattened equivalents only if agent.session_id path missing:
		// belt-and-suspenders string check then fail with guidance.
		if strings.Contains(out, fixtureGrokSessionID) && strings.Contains(out, "session_id") {
			t.Fatalf("session id present but not under agent.session_id path; fix JSON shape:\n%s", out)
		}
		t.Fatalf("no session with agent.session_id=%q:\n%s", fixtureGrokSessionID, out)
	}
	if foundKind != "" && foundKind != "grok" {
		t.Fatalf("agent.kind=%q want grok", foundKind)
	}
	if treeLen == 0 {
		// Soft preference logged as hard for P4 tree-in-JSON contract.
		t.Fatalf("agent.tree empty or missing; want busyGrokResolve nodes:\n%s", out)
	}
}
```
