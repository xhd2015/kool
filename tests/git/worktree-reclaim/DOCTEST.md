# worktree-reclaim

`kool git worktree reclaim` removes stale linked worktrees when they are clean and their HEAD is already contained in the main repo's current HEAD.

# DSN (Domain Specific Notion)

The **user** invokes `kool git worktree reclaim` with either a linked worktree path or `--all` to process every linked worktree of the current repository. The **reclaim handler** resolves the main repository (from the worktree `.git` gitdir or from the current working directory), lists linked worktrees, and evaluates each candidate. The **git layer** checks cleanliness via `git status --porcelain` and inclusion via branch comparison against main `HEAD`. When reclaimable and not dry-run, the handler runs `git worktree remove` and deletes the associated branch. The handler **prints** per-worktree outcomes to stdout and returns an exit code reflecting errors.

## Version

0.0.2

## Decision Tree

```
mode
├── reclaim-dir                         # kool git worktree reclaim <dir>
│   ├── clean-merged/
│   │   ├── leaf                        # clean + HEAD included → removed + branch deleted
│   │   ├── dry-run/leaf                # --dry-run, worktree still exists
│   │   └── detached-head/leaf          # detached HEAD, commit included in main
│   ├── dirty/leaf                      # uncommitted changes → error, not removed
│   ├── unmerged-ahead/leaf             # branch ahead of main HEAD → error
│   ├── diverged/leaf                   # diverged from main HEAD → error
│   ├── not-a-worktree/leaf             # path is main repo or non-git → error
│   ├── invalid-path/leaf               # path does not exist → error
│   └── dead-worktree/leaf              # dir deleted, still in git list → reclaimed (dead)
└── reclaim-all                         # kool git worktree reclaim --all
    ├── mixed/leaf                      # one reclaimable, one dirty → reclaim one, skip one
    ├── none-reclaimable/leaf           # all skipped, exit 0
    ├── all-reclaimable/leaf            # all removed
    ├── dry-run/leaf                    # --dry-run: reports would-reclaim, nothing removed
    ├── from-linked-cwd/leaf            # run from inside a linked worktree cwd
    └── dead-worktree/
        ├── dry-run/leaf                # --dry-run: would reclaim (dead), registration remains
        └── leaf                        # dir deleted, still in git list → reclaimed (dead)
```

## Test Cases

| # | Path | Description |
|---|------|-------------|
| 1 | reclaim-dir/clean-merged/leaf | Clean merged worktree is reclaimed and branch deleted |
| 2 | reclaim-dir/clean-merged/dry-run/leaf | Dry-run reports would-reclaim; worktree remains |
| 3 | reclaim-dir/clean-merged/detached-head/leaf | Detached HEAD worktree with included commit is reclaimed |
| 4 | reclaim-dir/dirty/leaf | Uncommitted changes cause error; worktree remains |
| 5 | reclaim-dir/unmerged-ahead/leaf | Branch ahead of main causes error; worktree remains |
| 6 | reclaim-dir/diverged/leaf | Diverged branch causes error; worktree remains |
| 7 | reclaim-dir/not-a-worktree/leaf | Main repo path is rejected as not a linked worktree |
| 8 | reclaim-dir/invalid-path/leaf | Nonexistent path returns error |
| 9 | reclaim-all/mixed/leaf | Mixed candidates: reclaim one, skip one, exit 0 |
| 10 | reclaim-all/none-reclaimable/leaf | All worktrees skipped, exit 0 |
| 11 | reclaim-all/all-reclaimable/leaf | All reclaimable worktrees removed |
| 12 | reclaim-all/dry-run/leaf | Dry-run reports all would-reclaim; none removed |
| 13 | reclaim-all/from-linked-cwd/leaf | --all from linked worktree cwd resolves main repo |
| 14 | reclaim-dir/dead-worktree/leaf | Dead linked worktree reclaimed by path with (dead) message |
| 15 | reclaim-all/dead-worktree/dry-run/leaf | --all --dry-run reports would reclaim (dead); registration remains |
| 16 | reclaim-all/dead-worktree/leaf | Dead linked worktree reclaimed via --all with (dead) message |

## How to Run

```sh
doctest vet ./tests/git/worktree-reclaim
doctest test ./tests/git/worktree-reclaim
```

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

type Request struct {
	All    bool   // reclaim --all
	Path   string // linked worktree path for single reclaim
	DryRun bool
	Cwd    string // working directory for kool invocation; empty uses main repo

	// Populated by Setup helpers:
	MainRepo     string
	WorktreePath string
	BranchName   string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := []string{"git", "worktree", "reclaim"}
	if req.All {
		args = append(args, "--all")
	} else {
		args = append(args, req.Path)
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command("kool", args...)
	if req.Cwd != "" {
		cmd.Dir = req.Cwd
	} else if req.MainRepo != "" {
		cmd.Dir = req.MainRepo
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run kool: %w", runErr)
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```