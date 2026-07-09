# Scenario

**Feature**: empty previous title is printed as-is in the success line

```
# mock get returns empty string
kool iterm2 set-title fresh-name
  -> stdout "title changed:  -> fresh-name\n"
```

## Steps

1. Mock osascript stdout as empty (old title empty).
2. Set a concrete new title.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Title = "fresh-name"
	req.TitleSet = true
	req.OsascriptStdout = ""
	// Force empty mock: fake only prints when env non-empty; empty old is "".
	// Leaves that need empty mock leave OsascriptStdout unset/empty intentionally.
	return nil
}
```
