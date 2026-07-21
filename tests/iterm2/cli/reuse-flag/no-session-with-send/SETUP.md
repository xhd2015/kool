# Scenario

**Feature**: `-r` miss path with `--send` follow-ups only in else branch

```
kool iterm2 -r <dir> --send grok -> else branch: cd + write text "grok"
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