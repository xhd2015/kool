# Scenario

**Feature**: set-title surfaces osascript failure as exit 1

```
# in-session valid title, fake osascript fails
kool iterm2 set-title ok-title
  -> osascript exit 1 -> CLI exit 1
```

## Steps

1. In-session with a non-empty title.
2. Force fake osascript non-zero exit.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InSession = true
	req.Title = "ok-title"
	req.TitleSet = true
	req.OsascriptStdout = "old-title"
	req.OsascriptExit = 1
	return nil
}
```
