# Scenario

**Feature**: set-title with an empty string title is rejected

```
# empty title positional
kool iterm2 set-title ""
  -> stderr validation error, exit 1
```

## Steps

1. In-session so validation (not not-in-iTerm2) applies.
2. Pass empty string as the title positional (`TitleSet=true`, `Title=""`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InSession = true
	req.Title = ""
	req.TitleSet = true
	return nil
}
```
