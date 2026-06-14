## Steps
- Create branch `a` from main
- On branch `a`: modify foo.txt to "world" and commit
- On `main`: modify foo.txt to "world" (same content, different commit) and commit
- Now main and a are divergent but foo.txt has the same content on both sides
- The three-dot diff from merge-base shows 0 files difference
- Set RefA to `main` and RefB to `a`

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
	err := os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("world"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "change on a to world")

	runGit("checkout", "main")
	err = os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("world"), 0644)
	if err != nil {
		return err
	}
	runGit("add", ".")
	runGit("commit", "-m", "change on main to world")

	req.RefA = "main"
	req.RefB = "a"
	return nil
}
```
