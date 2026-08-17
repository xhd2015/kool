# Scenario

**Feature**: HandleWith `GOROOT=<abs>` skips resolve and does not call Install

```
# remaining cmd is /usr/bin/true so ExecGoroot needs no real SDK
GOROOT=/abs/fake /usr/bin/true -> HandleWith -> ExecGoroot; Install unused
```

## Steps

1. Set args to `GOROOT=<abs fake>` plus `/usr/bin/true`.
2. Set `Download` true so a mistaken resolve would call the hook.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	fake := filepath.Join(t.TempDir(), "fake-sdk")
	req.Args = []string{"GOROOT=" + fake, "/usr/bin/true"}
	req.Download = true
	req.Prompt = "should-not-print\n"
	return nil
}
```
