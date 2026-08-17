# Scenario

**Feature**: get-title prints the current session title

```
# mock osascript stdout = Tab Alpha
kool iterm2 get-title
  -> stdout "Tab Alpha\n"
  -> script targets session name for UUID
```

## Steps

1. Mock osascript stdout `Tab Alpha` (no trailing newline in mock; CLI adds `\n`).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OsascriptStdout = "Tab Alpha"
	return nil
}
```
