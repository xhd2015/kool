## Expected

- Exit code 0
- Dry-run merge command references branch `feature`, not a raw commit hash

## Side Effects

- No git mutations

## Exit Code

- 0

```go
import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var fullCommitHash = regexp.MustCompile(`[0-9a-f]{40}`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	out := combinedOutput(resp)
	wantMerge := fmt.Sprintf("merge --ff-only %s", req.BranchName)
	if !strings.Contains(out, wantMerge) {
		t.Fatalf("expected dry-run to use branch name in merge command %q, got:\n%s", wantMerge, out)
	}
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "merge --ff-only") {
			continue
		}
		if fullCommitHash.MatchString(line) {
			t.Fatalf("attached worktree must not use commit hash in merge command, got line:\n%s\nfull output:\n%s", line, out)
		}
	}
}
```