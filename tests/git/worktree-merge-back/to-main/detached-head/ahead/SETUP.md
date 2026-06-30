# Scenario

**Feature**: detached HEAD commit is ahead of main HEAD

```
# worktree commit not reachable from main; must not be classified as already-included
merge-back handler -> compare detached commit vs main HEAD -> ahead -> confirmation required
```

## Context

- `ReadBranch` returns `HEAD` when detached; comparison must use the worktree commit SHA, not the literal ref name `HEAD` in the target repo

```go
import (
	"os/exec"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorktreePath == "" {
		t.Fatal("expected detached worktree from ancestor setup")
	}
	cmd := exec.Command("git", "-C", req.WorktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD in worktree: %v", err)
	}
	if strings.TrimSpace(string(out)) != "HEAD" {
		t.Fatalf("expected detached HEAD worktree, got branch %q", strings.TrimSpace(string(out)))
	}
	if fileTrackedInRepo(t, req.MainRepo, "detached-ahead.txt") {
		t.Fatal("main must not contain detached commit before merge-back")
	}
	return nil
}
```