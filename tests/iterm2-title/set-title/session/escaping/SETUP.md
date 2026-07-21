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
import "testing"

func Setup(t *testing.T, req *Request) error {
	markSetTitleSessionTree()
	markRootTree()
	markSessionTree()
	markSetTitleTree()
	req.Title = `say "hi"\path`
	req.TitleSet = true
	req.OsascriptStdout = "prev"
	return nil
}
```
