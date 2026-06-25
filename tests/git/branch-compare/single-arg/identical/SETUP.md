# Scenario

**Feature**: single-arg when refA equals current branch reports identical

```
# on main; user compares main against implicit refB=main
user (on main) -> kool git compare-branch main -> compare_branch.Handle

# refB defaults to main
compare_branch.Handle -> git branch --show-current -> refB=main
```

## Steps

- Create a temporary git repository checked out on `main`
- Set RefA to `main` and leave RefB empty

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir, err := os.MkdirTemp("", "kool-branch-compare-single-identical-*")
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

	req.Dir = dir
	req.RefA = "main"
	req.RefB = ""
	return nil
}
```