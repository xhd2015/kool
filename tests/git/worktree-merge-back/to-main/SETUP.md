# Scenario

**Feature**: merge-back targets the main repository by default

```
# default target resolves to main repo of source worktree
user (cwd=linked wt) -> merge-back handler -> target = main repo
```

## Context

- No `--to` flag; target is the main repository linked to the source worktree

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.To = ""
	req.TargetPath = ""
	return nil
}
```