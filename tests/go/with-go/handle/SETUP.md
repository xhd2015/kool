# Scenario

**Feature**: HandleWith dispatches empty args, GOROOT=, or version then ExecGoroot

```
# empty -> error; GOROOT= skips resolve; go1.19 -> ResolveGorootWith then ExecGoroot
args -> HandleWith -> error | ExecGoroot(GOROOT) | ResolveGorootWith then ExecGoroot
```

## Steps

1. Set `req.Op` to `handle`.
2. Inject `InstallDir` (`t.TempDir()`) and a recording Install hook.
3. Leaf Setup sets Args / dest / fake go.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "handle"
	req.InstallDir = t.TempDir()
	req.RecordInstall = true
	return nil
}
```
