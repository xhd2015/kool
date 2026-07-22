# Scenario

**Feature**: vscode tasks show — exact or unique CI substring

```
show <label> -> task details | missing/ambiguous error
```

## Steps

1. Subcommand `show`.
2. Leaves set Query and fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "show"
	return nil
}
```
