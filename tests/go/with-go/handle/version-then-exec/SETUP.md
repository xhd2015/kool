# Scenario

**Feature**: HandleWith `go1.19` resolves dest then ExecGoroot of fake `bin/go`

```
# dest exists so resolve does not install; bare go looks up $GOROOT/bin/go
go1.19 go -> HandleWith -> ResolveGorootWith(dest-exists) -> ExecGoroot(fake go)
```

## Steps

1. Create dest `$InstallDir/go1.19.13` with a fake `bin/go`.
2. Set args `go1.19` `go`. Download true so a missing dest would call the hook.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	dest := destPin(req.InstallDir)
	writeFakeGo(t, dest)
	req.Args = []string{"go1.19", "go"}
	req.Download = true
	req.Prompt = "should-not-print\n"
	return nil
}
```
