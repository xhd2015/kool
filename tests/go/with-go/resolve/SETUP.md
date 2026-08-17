# Scenario

**Feature**: ResolveGorootWith returns dest `$InstallDir/<pin>` or installs via hook

```
# pin go1.19 -> dest=$InstallDir/go1.19.13
goVersion -> PinPatch -> dest exists? return dest : Download? Install hook : error
```

## Steps

1. Set `req.Op` to `resolve`, `GoVersion` to `go1.19`.
2. Inject `InstallDir` (`t.TempDir()`) and a recording Install hook.
3. Child Setup creates dest or sets Download / Prompt / HookGoroot.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "resolve"
	req.GoVersion = "go1.19"
	req.InstallDir = t.TempDir()
	req.RecordInstall = true
	return nil
}
```
