# Scenario

**Feature**: tab-set show prints one config

```
tab-set show <name> -> window_name + tabs id/command
```

## Steps

1. Subcommand `show`; leaves set SetName and fixtures.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Subcommand = "show"
	return nil
}
```
