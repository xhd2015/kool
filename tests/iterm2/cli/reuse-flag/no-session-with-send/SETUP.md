# Scenario

**Feature**: `-r` miss path with `--send` follow-ups only in else branch

```
kool iterm2 -r <dir> --send grok -> else branch: cd + write text "grok"
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