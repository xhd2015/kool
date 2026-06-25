# Scenario

**Feature**: single-arg with -C flag defaults refB inside the target repo

```
# repo in subdir; user compares main from outside cwd via -C
user -> kool git compare-branch main -C <repoDir> -> compare_branch.Handle

# refB resolved inside repoDir current branch (feature)
compare_branch.Handle -> git branch --show-current (in repoDir) -> refB=feature
```

## Steps

- Create a git repository in a subdirectory under a temp working directory
- On `main`, add one commit ahead of `feature`; check out `feature`
- Set RefA to `main`, leave RefB empty, and set Dir to the repository path

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	workDir, err := os.MkdirTemp("", "kool-branch-compare-single-C-*")
	if err != nil {
		return err
	}
	t.Cleanup(func() { os.RemoveAll(workDir) })

	repoDir := filepath.Join(workDir, "repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		return err
	}

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "test")

	err = os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# test"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "initial commit")
	runGit("branch", "-M", "main")

	runGit("checkout", "-b", "feature")
	runGit("checkout", "main")
	err = os.WriteFile(filepath.Join(repoDir, "file1.txt"), []byte("content 1"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "commit on main")

	runGit("checkout", "feature")

	req.Dir = repoDir
	req.RefA = "main"
	req.RefB = ""
	return nil
}
```