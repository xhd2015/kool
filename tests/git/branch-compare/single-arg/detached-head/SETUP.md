# Scenario

**Feature**: single-arg in detached HEAD defaults refB to HEAD

```
# detached at feature tip; user compares main against implicit HEAD
user (detached HEAD) -> kool git compare-branch main -> compare_branch.Handle

# empty branch name -> refB=HEAD
compare_branch.Handle -> git branch --show-current (empty) -> refB=HEAD
```

## Steps

- Create a temporary git repository with `main` ahead of `feature`
- Detach HEAD at the `feature` commit
- Set RefA to `main` and leave RefB empty

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir, err := os.MkdirTemp("", "kool-branch-compare-single-detached-*")
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

	runGit("checkout", "--detach", "feature")

	req.Dir = dir
	req.RefA = "main"
	req.RefB = ""
	return nil
}
```