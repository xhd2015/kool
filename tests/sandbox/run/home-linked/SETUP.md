# Scenario

**Feature**: sealed binary with home_linked seeds real home and sets guest HOME

```
# home-linked sealed run
kool sandbox build -o sandbox.bin --home-linked …
HOME=FAKE_REAL_HOME KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- <command>
  -> capture real home from HOME; seed top-level links; overlay pack;
     guest HOME=SANDBOX_ROOT=session; exec command
```

## Steps

1. Enable `HomeLinked` on build.
2. Ensure a fake real home directory under WorkingDir and set `SealedHome`
   so the sealed process receives `HOME=<fake>` (child env only).
3. AfterBuildRun remains true (from run/ parent).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HomeLinked = true
	if req.SealedHome == "" {
		home := filepath.Join(req.WorkingDir, "fake-real-home")
		if err := os.MkdirAll(home, 0755); err != nil {
			return err
		}
		req.SealedHome = home
	}
	return nil
}
```
