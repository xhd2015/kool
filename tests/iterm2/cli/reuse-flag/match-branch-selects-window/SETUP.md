# Scenario

**Feature**: `-r` match branch brings matchingWindow to front before focusing tab/session

```
kool iterm2 -r <dir> -> match branch: select matchingWindow -> select tab/session (not background window only)
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