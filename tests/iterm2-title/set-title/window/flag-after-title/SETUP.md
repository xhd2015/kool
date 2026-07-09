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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WindowAfterTitle = true
	req.Title = "new-window-title"
	req.OsascriptStdout = "old-window-title"
	return nil
}
```
