# Scenario

**Feature**: vscode tasks list reads workspace tasks.json

```
--dir | cwd -> walk-up .vscode/tasks.json -> list table / JSON
```

## Steps

1. Subcommand `list`.
2. Leaves prepare fixtures or empty sandbox.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "list"
	return nil
}
```
