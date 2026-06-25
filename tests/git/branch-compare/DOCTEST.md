# compare-branch

`kool git compare-branch` compares two git references and reports whether they are identical, fast-forwardable, or divergent.

# DSN (Domain Specific Notion)

The **user** invokes the kool CLI with one or two git refs and an optional repository directory. The **compare-branch handler** parses positional refs and the `-C` flag, defaulting the second ref to the current branch name (or `HEAD` when detached) when only one ref is given. The **git layer** resolves both refs inside the target repository and computes their commit relationship. The handler **prints** a human-readable report to stdout describing identical commits, fast-forward direction, or divergence.

## Version

0.0.2

## Decision Tree

```
arg-count
├── single-arg                    # refB defaults to current branch or HEAD
│   ├── on-branch/leaf            # checked out on named branch feature
│   ├── detached-head/leaf        # detached HEAD state
│   ├── identical/leaf            # refA equals current branch
│   └── with-C-flag/leaf          # single arg with -C <dir>
└── two-arg                       # explicit refA and refB
    ├── identical/leaf
    ├── fast-forward/
    │   ├── a-to-b
    │   └── b-to-a
    ├── divergent/
    │   ├── leaf
    │   └── no-file-diff/leaf
    ├── errors/
    │   ├── invalid-ref-a
    │   └── invalid-ref-b
    ├── with-C-flag/
    │   ├── valid-dir
    │   └── nonexistent
    └── not-a-git-repo/leaf
```

## Test Cases

| # | Path | Description |
|---|------|-------------|
| 1 | single-arg/on-branch/leaf | One arg on named branch; refB defaults to current branch |
| 2 | single-arg/detached-head/leaf | One arg in detached HEAD; refB defaults to HEAD |
| 3 | single-arg/identical/leaf | One arg matching current branch; identical report |
| 4 | single-arg/with-C-flag/leaf | One arg with -C pointing at repo |
| 5 | identical/leaf | Two refs resolve to the same commit |
| 6 | fast-forward/a-to-b | a is ancestor of b |
| 7 | fast-forward/b-to-a | b is ancestor of a |
| 8 | divergent/leaf | Both have unique commits with file differences |
| 9 | divergent/no-file-diff/leaf | Divergent commits, zero file differences |
| 10 | errors/invalid-ref-a | refA does not resolve |
| 11 | errors/invalid-ref-b | refB does not resolve |
| 12 | with-C-flag/valid-dir | -C points to a valid git repo |
| 13 | with-C-flag/nonexistent | -C points to a non-existent directory |
| 14 | not-a-git-repo/leaf | Directory is not a git repository |

## How to Run

```sh
doctest vet ./tests/git/branch-compare
doctest test ./tests/git/branch-compare
```

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

type Request struct {
	RefA string
	RefB string // empty means single-arg CLI invocation (only RefA passed)
	Dir  string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	args := []string{"git", "compare-branch", req.RefA}
	if req.RefB != "" {
		args = append(args, req.RefB)
	}
	if req.Dir != "" {
		args = append(args, "-C", req.Dir)
	}
	cmd := exec.Command("kool", args...)
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