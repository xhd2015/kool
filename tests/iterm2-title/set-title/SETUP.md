# Scenario

**Feature**: set-title subcommand changes session or window title

```
# set-title pipeline
kool iterm2 set-title [--window] <title> + ITERM_SESSION_ID
  -> get old title -> set new title -> stdout "title changed: old -> new"
```

## Steps

1. Fix `Command` to `set-title` for all descendants.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Command = "set-title"
	return nil
}
```
