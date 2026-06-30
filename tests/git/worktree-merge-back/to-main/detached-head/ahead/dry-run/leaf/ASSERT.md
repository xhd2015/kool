## Expected

- Exit code 0
- Dry-run lists a fast-forward merge plan; does not claim already-included

## Side Effects

- Main repo does not contain `detached-ahead.txt`
- Worktree unchanged

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	out := combinedOutput(resp)
	if strings.Contains(strings.ToLower(out), "already included") {
		t.Fatalf("dry-run must not report detached ahead commit as already included, got:\n%s", out)
	}
	commit := revParseHEAD(t, req.WorktreePath)
	wantMerge := "merge --ff-only " + commit
	if !strings.Contains(out, wantMerge) {
		t.Fatalf("detached HEAD must use commit hash in merge command %q, got:\n%s", wantMerge, out)
	}
	if strings.Contains(out, "merge --ff-only HEAD") {
		t.Fatalf("detached HEAD must not use symbolic HEAD ref in merge command, got:\n%s", out)
	}
	if strings.Contains(out, "merge --ff-only "+req.BranchName) {
		t.Fatalf("detached HEAD must not use branch name %q when not checked out, got:\n%s", req.BranchName, out)
	}
	if fileTrackedInRepo(t, req.MainRepo, "detached-ahead.txt") {
		t.Fatalf("dry-run must not merge detached commit into main")
	}
}
```