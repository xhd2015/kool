# Scenario

**Feature**: --window may appear after the title positional

```
# flag after title (lessflags order flexibility)
kool iterm2 set-title new-window-title --window
  -> same success as --window before title
```

## Steps

1. Set `WindowAfterTitle=true` so argv is `set-title <title> --window`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.WindowAfterTitle = true
	req.Title = "new-window-title"
	req.OsascriptStdout = "old-window-title"
	return nil
}
```
