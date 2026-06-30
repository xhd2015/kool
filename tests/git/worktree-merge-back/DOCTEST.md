# worktree-merge-back

`kool git worktree merge-back` merges a linked worktree branch into a target checkout (main repo or sibling worktree), optionally rebasing when histories diverge, with interactive confirmation for mutating operations.

# DSN (Domain Specific Notion)

The **user** runs `kool git worktree merge-back` from inside a linked worktree (or with flags that identify source and target). The **merge-back handler** validates that the cwd is a linked worktree, resolves the target path (default main repo or `--to`), and checks cleanliness. The **git layer** compares the source branch against the target `HEAD` to classify the relationship as already-included, ahead, or diverged. For ahead and diverged cases the handler builds a **plan** of concrete `git -C` commands and requires **confirmation** on a TTY (or `--confirm-from-stdin` for scripted input); non-interactive stdin without that flag is rejected. On approval the handler executes fast-forward merge and/or rebase, then optionally removes the worktree and deletes the branch when `--rm` is set. With `--dry-run` the handler prints the planned commands without mutating git state.

## Version

0.0.2

## Decision Tree

```
outcome
├── validation/                         # preflight rejects invalid invocation
│   ├── not-a-worktree/leaf             # cwd is main repo → error
│   ├── dirty/leaf                      # uncommitted changes → error
│   ├── same-target/leaf                # --to same path as source → error
│   └── foreign-target/leaf             # --to worktree from different main repo → error
├── to-main/                            # default target = main repo
│   ├── already-included/
│   │   ├── noop/leaf                   # included, no --rm → success, worktree remains
│   │   └── with-rm/leaf                # included + --rm → worktree removed + branch deleted
│   ├── detached-head/
│   │   └── ahead/
│   │       ├── leaf                    # detached HEAD ahead of main → merge, not already-included
│   │       └── dry-run/leaf            # dry-run lists merge plan, not already-included
│   ├── ahead/
│   │   ├── non-tty/leaf                # no TTY → error, no mutations
│   │   ├── confirm-decline/leaf        # confirm-from-stdin + 'n' → abort exit 0
│   │   ├── confirm-default/leaf        # confirm-from-stdin + Enter → ff merge, worktree remains
│   │   ├── merge-and-rm/leaf           # confirm + --rm → merged + removed
│   │   └── dry-run/
│   │       ├── leaf                    # --dry-run prints planned commands, no changes
│   │       └── merge-uses-branch/leaf  # dry-run uses branch name, not commit hash
│   └── diverged/
│       ├── non-tty/leaf                # no TTY → error
│       ├── rebase-fails/leaf           # confirm → rebase conflict → error
│       ├── rebase-succeeds/leaf        # confirm → rebase + ff merge
│       └── rebase-succeeds-rm/leaf     # confirm + --rm
└── to-worktree/                        # --to sibling worktree
    ├── ahead/leaf                      # merge into sibling worktree HEAD
    └── dry-run/leaf                    # --dry-run with --to sibling
```

## Test Cases

| # | Path | Description |
|---|------|-------------|
| 1 | validation/not-a-worktree/leaf | Main repo cwd rejected as not a linked worktree |
| 2 | validation/dirty/leaf | Uncommitted changes cause error; worktree unchanged |
| 3 | validation/same-target/leaf | `--to` same as source worktree is rejected |
| 4 | validation/foreign-target/leaf | `--to` worktree from foreign main repo is rejected |
| 5 | to-main/already-included/noop/leaf | Already-included branch is no-op; worktree kept |
| 6 | to-main/already-included/with-rm/leaf | Already-included + `--rm` removes worktree and branch |
| 7 | to-main/detached-head/ahead/leaf | Detached HEAD ahead of main merges; not falsely already-included |
| 8 | to-main/detached-head/ahead/dry-run/leaf | Detached HEAD dry-run lists merge; not falsely already-included |
| 9 | to-main/ahead/non-tty/leaf | Ahead branch without TTY confirmation is rejected |
| 10 | to-main/ahead/confirm-decline/leaf | User declines merge; no git mutations |
| 11 | to-main/ahead/confirm-default/leaf | User confirms; target ff-merged; source worktree remains |
| 12 | to-main/ahead/merge-and-rm/leaf | User confirms with `--rm`; merged and worktree removed |
| 13 | to-main/ahead/dry-run/leaf | Dry-run lists planned commands; no mutations |
| 14 | to-main/ahead/dry-run/merge-uses-branch/leaf | Attached branch dry-run uses branch name in merge |
| 15 | to-main/diverged/non-tty/leaf | Diverged branch without TTY confirmation is rejected |
| 16 | to-main/diverged/rebase-fails/leaf | Rebase conflict aborts; source unchanged |
| 17 | to-main/diverged/rebase-succeeds/leaf | Rebase + ff merge succeeds into main |
| 18 | to-main/diverged/rebase-succeeds-rm/leaf | Rebase + merge + `--rm` removes worktree |
| 19 | to-worktree/ahead/leaf | Ahead branch merged into sibling worktree HEAD |
| 20 | to-worktree/dry-run/leaf | Dry-run with `--to` sibling lists commands only |

## How to Run

```sh
doctest vet ./tests/git/worktree-merge-back
doctest test ./tests/git/worktree-merge-back
```

```go
import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"testing"
)

type Request struct {
	To               string // --to target path; empty means main repo
	DryRun           bool
	Remove           bool
	ConfirmFromStdin bool
	StdinInput       string
	Cwd              string // working directory for kool; defaults to source worktree

	// Populated by Setup helpers:
	MainRepo     string
	WorktreePath string
	TargetPath   string
	SiblingPath  string
	BranchName   string
	ForeignMain  string
	ForeignWT    string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := []string{"git", "worktree", "merge-back"}
	if req.To != "" {
		args = append(args, "--to", req.To)
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.Remove {
		args = append(args, "--rm")
	}
	if req.ConfirmFromStdin {
		args = append(args, "--confirm-from-stdin")
	}

	cwd := req.Cwd
	if cwd == "" && req.WorktreePath != "" {
		cwd = req.WorktreePath
	}

	cmd := exec.Command("kool", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	if req.StdinInput != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start kool: %w", err)
		}
		if _, err := io.WriteString(stdin, req.StdinInput); err != nil {
			return nil, err
		}
		stdin.Close()
		runErr := cmd.Wait()
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