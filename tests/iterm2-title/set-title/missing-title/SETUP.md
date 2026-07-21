# Scenario

**Feature**: set-title without a title argument is a validation error

```
# missing title (in session so we pass the iTerm2 gate)
kool iterm2 set-title
  -> stderr validation error, exit 1
```

## Steps

1. Set `InSession=true` so the failure is missing title, not not-in-iTerm2.
2. Leave `TitleSet=false` (no title positional).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	markSetTitleTree()
	req.InSession = true
	req.TitleSet = false
	return nil
}
```
