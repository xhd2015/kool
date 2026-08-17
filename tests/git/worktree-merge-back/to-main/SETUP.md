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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.To = ""
	req.TargetPath = ""
	return nil
}
```