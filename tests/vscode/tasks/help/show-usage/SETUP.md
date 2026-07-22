# Scenario

**Feature**: tasks --help prints usage for list/find/show/run

```
vscode tasks --help
  -> exit 0; stdout mentions tasks, list, find, show, run, dry-run
```

## Steps

1. Request Help on vscode tasks.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	return nil
}
```
