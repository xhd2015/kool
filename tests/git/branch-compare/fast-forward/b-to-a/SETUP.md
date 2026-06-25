# Scenario

**Feature**: refB can fast-forward to refA (b-to-a direction)

```
# main (refA) is one commit ahead of a (refB)
compare_branch.Handle(refA=main, refB=a) -> a fast-forwards to main
```

## Steps
- Create branch `a` from the initial commit
- Add one commit on `main` (so main is ahead of a)
- Set RefA to `main` and RefB to `a` — b (a) can fast-forward to a (main)

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

	req.RefA = "main"
	req.RefB = "a"
	return nil
}
```
