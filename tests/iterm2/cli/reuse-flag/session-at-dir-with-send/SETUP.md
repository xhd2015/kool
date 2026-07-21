# Scenario

**Feature**: `-r` match path suppresses `--send` follow-ups

```
kool iterm2 -r <dir> --send grok -> match branch: focus only (no write text "grok")
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markCliReuseFlagTree()
	markCliTree()
	markRootTree()
	req.Send = []string{"grok"}
	return nil
}
```