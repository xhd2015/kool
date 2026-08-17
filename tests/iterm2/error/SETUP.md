# Scenario

**Feature**: CLI surfaces install, osascript, and platform errors

```
kool iterm2 -> OpenConfig failure -> stderr + non-zero exit
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Phase == "" {
		req.Phase = "cli"
	}
	if req.InstalledEnv == "" {
		req.InstalledEnv = "1"
	}
	return nil
}
```