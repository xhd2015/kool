# Scenario

**Feature**: `-r` match path suppresses `--send` follow-ups

```
kool iterm2 -r <dir> --send grok -> match branch: focus only (no write text "grok")
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Send = []string{"grok"}
	return nil
}
```