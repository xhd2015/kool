# Scenario

**Feature**: vscode tasks list reads workspace tasks.json

```
--dir | cwd -> walk-up .vscode/tasks.json -> list table / JSON
```

## Steps

1. Subcommand `list`.
2. Leaves prepare fixtures or empty sandbox.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Subcommand = "list"
	return nil
}
```
