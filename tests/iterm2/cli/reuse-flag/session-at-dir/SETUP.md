# Scenario

**Feature**: `-r` match path — focus existing session/tab only

```
kool iterm2 -r <dir> -> scan finds path == targetDir -> focus session/tab (no cd, no tab create)
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