# Scenario

**Feature**: single-arg identical refs on current branch

```
# refA and defaulted refB both resolve to main
compare_branch.Handle(refA=main, refB=main) -> identical report
```

## Steps

- Verify the repository is checked out on `main` with RefA=`main` and empty RefB

```go
import (
	"os/exec"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = req.Dir
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != "main" {
		t.Fatalf("expected current branch main, got %q", strings.TrimSpace(string(out)))
	}
	if req.RefA != "main" || req.RefB != "" {
		t.Fatalf("unexpected refs: RefA=%q RefB=%q", req.RefA, req.RefB)
	}
	return nil
}
```