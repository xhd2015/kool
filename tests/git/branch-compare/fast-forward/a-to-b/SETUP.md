# Scenario

**Feature**: refA can fast-forward to refB (a-to-b direction)

```
# a is ancestor; main (refB) is one commit ahead
compare_branch.Handle(refA=a, refB=main) -> a fast-forwards to main
```

## Steps
- Create branch `a` from the initial commit
- Add one commit on `main` (so main is ahead of a)
- Set RefA to `a` and RefB to `main` — a can fast-forward to main

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := req.Dir
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("checkout", "-b", "a")
	runGit("checkout", "main")
	err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content 1"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "commit on main")

	req.RefA = "a"
	req.RefB = "main"
	return nil
}
```
