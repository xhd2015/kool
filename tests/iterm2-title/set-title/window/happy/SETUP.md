# Scenario

**Feature**: set-title --window happy path

```
kool iterm2 set-title --window new-window-title
  -> title changed: old-window-title -> new-window-title
  -> script sets window name for the session's window
```

## Steps

1. Window target; mock old `old-window-title`; new `new-window-title`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Title = "new-window-title"
	req.TitleSet = true
	req.OsascriptStdout = "old-window-title"
	return nil
}
```
