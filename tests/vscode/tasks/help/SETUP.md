# Scenario

**Feature**: kool vscode tasks help

```
kool vscode tasks -h|--help -> usage on stdout exit 0
```

## Steps

1. Leaves set Help=true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping: help leaves set Help=true; clear subcommand.
	req.Subcommand = ""
	return nil
}
```
