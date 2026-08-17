# Scenario

**Feature**: get-title --window prints the window title

```
# mock osascript stdout = Project Window
kool iterm2 get-title --window
  -> stdout "Project Window\n"
```

## Steps

1. Mock osascript stdout `Project Window`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OsascriptStdout = "Project Window"
	return nil
}
```
