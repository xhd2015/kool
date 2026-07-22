# Scenario

**Feature**: vscode tasks find — case-insensitive substring on label

```
find <query> -> matching task rows | error if zero
```

## Steps

1. Subcommand `find`.
2. Leaves set Query and fixture.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "find"
	return nil
}
```
