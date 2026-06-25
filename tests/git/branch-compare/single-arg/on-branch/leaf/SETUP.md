# Scenario

**Feature**: single-arg on named branch produces same output as two-arg main feature

```
# implicit refB=feature; main is one commit ahead
compare_branch.Handle(refA=main, refB=feature) -> fast-forward report
```

## Steps

- Verify the repository is checked out on `feature` with RefA=`main` and empty RefB

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
		t.Fatalf("expected current branch feature, got %q", strings.TrimSpace(string(out)))
	}
	if req.RefA != "main" || req.RefB != "" {
		t.Fatalf("unexpected refs: RefA=%q RefB=%q", req.RefA, req.RefB)
	}
	return nil
}
```