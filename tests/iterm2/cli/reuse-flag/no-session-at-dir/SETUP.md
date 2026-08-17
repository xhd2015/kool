# Scenario

**Feature**: `-r` miss path — new window and cd in generated script

```
kool iterm2 -r <dir> -> ModeReuseCurrent script -> else branch: create window + cd
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Send = nil
	return nil
}
```