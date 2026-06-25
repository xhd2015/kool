# Scenario

**Feature**: -C points to a valid git repository

```
# repo in subdir; both refs resolve inside it
compare_branch.Handle(refA=main, refB=main, dir=repoDir) -> identical report
```

## Steps
- Create a separate git repository in a subdirectory
- Override req.Dir to point to that git repository
- Set RefA and RefB to `main` (both resolve to same commit)

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.Dir, "repo")
	err := os.MkdirAll(repoDir, 0755)
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

	req.Dir = repoDir
	req.RefA = "main"
	req.RefB = "main"
	return nil
}
```
