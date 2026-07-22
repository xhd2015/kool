# Scenario

**Feature**: tasks.json parse validation

```
invalid JSONC / broken file -> Error on list or any command
```

## Steps

1. Leaves write invalid fixtures; typically invoke list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// default exercise via list
	req.Subcommand = "list"
	return nil
}
```
