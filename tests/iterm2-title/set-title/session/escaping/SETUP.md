# Scenario

**Feature**: set-title escapes quotes and backslashes in the new title

```
# title contains " and \
kool iterm2 set-title 'say "hi"\path'
  -> success message includes the new title text
  -> AppleScript embeds escaped literals (\" and \\)
```

## Steps

1. Title value: `say "hi"\path` (Go string with quote and backslash).
2. Mock a simple old title for the success line.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Title = `say "hi"\path`
	req.TitleSet = true
	req.OsascriptStdout = "prev"
	return nil
}
```
