# Scenario

**Feature**: set-title happy path updates session title and prints old -> new

```
# mock get returns old-session-title
kool iterm2 set-title new-session-title
  -> stdout "title changed: old-session-title -> new-session-title\n"
  -> script references session UUID and sets session name
```

## Steps

1. In-session session target; mock old title `old-session-title`.
2. New title `new-session-title`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Title = "new-session-title"
	req.TitleSet = true
	req.OsascriptStdout = "old-session-title"
	return nil
}
```
