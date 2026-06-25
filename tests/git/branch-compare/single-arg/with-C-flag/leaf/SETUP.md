# Scenario

**Feature**: single-arg with -C produces same fast-forward result as on-branch case

```
# -C selects repo; implicit refB=feature inside that repo
compare_branch.Handle(refA=main, refB=feature, dir=repoDir) -> fast-forward report
```

## Steps

- Verify the target repo is checked out on `feature` with RefA=`main`, empty RefB, and Dir set

```go
import (
	"os/exec"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = req.Dir
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != "feature" {
		t.Fatalf("expected current branch feature in target repo, got %q", strings.TrimSpace(string(out)))
	}
	if req.RefA != "main" || req.RefB != "" || req.Dir == "" {
		t.Fatalf("unexpected request: RefA=%q RefB=%q Dir=%q", req.RefA, req.RefB, req.Dir)
	}
	return nil
}
```