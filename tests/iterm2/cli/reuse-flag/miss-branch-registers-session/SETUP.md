# Scenario

**Feature**: `-r` miss branch registers targetDir for immediate second-run reuse

After the first `kool iterm2 -r <dir>` opens a new window and sends `cd`, a second
immediate invocation must find that session. iTerm's read-only `path` variable may
lag behind `write text "cd …"`, so the miss branch must set a user session variable
that the scan can match on the next run.

```
kool iterm2 -r <dir> (no prior session) -> else branch: new window + cd + register user.koolTargetDir
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