# Scenario

**Feature**: single-arg on a named branch defaults refB to that branch name

```
# repo checked out on feature; user compares main against implicit refB
user (on feature) -> kool git compare-branch main -> compare_branch.Handle

# refB becomes feature via git branch --show-current
compare_branch.Handle -> git branch --show-current -> refB=feature
```

## Steps

- Create a temporary git repository with `main` and `feature` branches
- Add one commit on `main` so it is ahead of `feature`
- Check out `feature` (current branch)
- Set RefA to `main` and leave RefB empty

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir, err := os.MkdirTemp("", "kool-branch-compare-single-on-branch-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "test")

	err = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "initial commit")
	runGit("branch", "-M", "main")

	runGit("checkout", "-b", "feature")
	runGit("checkout", "main")
	err = os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content 1"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "commit on main")

	runGit("checkout", "feature")

	req.Dir = dir
	req.RefA = "main"
	req.RefB = ""
	return nil
}
```