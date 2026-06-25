# Scenario

**Feature**: detached HEAD single-arg uses HEAD in comparison output

```
# refB=HEAD at feature commit; main is ahead
compare_branch.Handle(refA=main, refB=HEAD) -> fast-forward report naming HEAD
```

## Steps

- Verify HEAD is detached with RefA=`main` and empty RefB

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
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("expected detached HEAD, got branch %q", strings.TrimSpace(string(out)))
	}
	if req.RefA != "main" || req.RefB != "" {
		t.Fatalf("unexpected refs: RefA=%q RefB=%q", req.RefA, req.RefB)
	}
	return nil
}
```