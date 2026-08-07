# Scenario

**Feature**: sealed runner merges load-devbox packs (files + env) at run time

```
# load-devbox runtime
kool sandbox build -o primary … [--runtime-load-devbox ABS]…
# secondary sealed binaries built under WorkingDir (SecondaryPacks)
KOOL_SANDBOX_ROOT=PARENT ./primary [--load-devbox ABS]... [--] <command>
  -> unseal primary; walk sealed then adhoc load paths; merge Files/Env;
     notice: loading devbox; materialize; exec guest
```

## Steps

1. Inherit run/ host GOOS build + AfterBuildRun + KOOL_SANDBOX_ROOT.
2. Leaves configure SecondaryPacks, SealedLoadDevbox, and/or primary RuntimeLoadDevbox.
3. Default SealedDoubleDash=true (runner flags then `--` then guest); validation /
   stop-on-first-arg may clear it.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Runner flags (--load-devbox) then -- then guest argv by default.
	req.SealedDoubleDash = true
	return nil
}
```
