# Scenario

**Feature**: `-r` smart-reuse AppleScript (scan paths, focus on match, new window on miss)

```
# -r uses same path scan as default smart-open
kool iterm2 -r <dir> [--send ...] -> shell/iterm2 ModeReuseCurrent -> AppleScript

# match: focus session/tab at targetDir — no cd, no follow-ups
# miss: new window + cd + optional --send lines
fake osascript <- captured script (both branches in one static script)
```

## Context

- Leaves assert structure of the generated AppleScript via substring checks on
  `resp.CapturedScript`; no live iTerm session state is simulated.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Reuse = true
	req.DirPath = initValidDir(t, req.WorkingDir, "reuse-target")
	return nil
}

func scriptHasReusePathScan(script string) bool {
	return strings.Contains(script, `variable named "path"`) &&
		(strings.Contains(script, "matchingWindow") ||
			strings.Contains(script, "matchingTab") ||
			strings.Contains(script, "matchingSession"))
}

func reuseScriptMatchBranch(script string) string {
	const open = `if matchingWindow is not missing value then`
	start := strings.Index(script, open)
	if start < 0 {
		return ""
	}
	rest := script[start+len(open):]
	elseIdx := strings.Index(rest, "\n  else\n")
	if elseIdx < 0 {
		elseIdx = strings.Index(rest, "\n  else")
	}
	if elseIdx < 0 {
		return rest
	}
	return rest[:elseIdx]
}

func reuseScriptElseBranch(script string) string {
	const marker = `create window with default profile`
	idx := strings.Index(script, marker)
	if idx < 0 {
		return ""
	}
	return script[idx:]
}

func matchBranchMustNotContain(t *testing.T, script, forbidden string) {
	t.Helper()
	branch := reuseScriptMatchBranch(script)
	if branch == "" {
		t.Fatalf("missing match branch in script: %q", script)
	}
	if strings.Contains(branch, forbidden) {
		t.Fatalf("match branch must not contain %q: branch=%q", forbidden, branch)
	}
}

func elseBranchMustContain(t *testing.T, script, needle string) {
	t.Helper()
	branch := reuseScriptElseBranch(script)
	if branch == "" {
		t.Fatalf("missing else (new window) branch in script: %q", script)
	}
	if !strings.Contains(branch, needle) {
		t.Fatalf("else branch missing %q: branch=%q", needle, branch)
	}
}
```