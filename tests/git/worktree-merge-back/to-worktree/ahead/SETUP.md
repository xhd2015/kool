# Scenario

**Feature**: merge ahead branch into sibling worktree

```
user -> merge-back --to sibling --confirm-from-stdin Enter -> sibling HEAD advanced
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ConfirmFromStdin = true
	req.StdinInput = "\n"
	req.Remove = false
	req.DryRun = false
	if req.SiblingPath == "" || req.To != req.SiblingPath {
		t.Fatal("expected sibling target from ancestor setup")
	}
	return nil
}
```