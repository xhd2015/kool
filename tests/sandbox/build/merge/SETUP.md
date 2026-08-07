# Scenario

**Feature**: merge input dir with flags; flags win on conflict

```
user -> kool sandbox build -o OUT -i DIR --file … --env …
  -> sealed OUT; inspect shows flag-winning paths/env keys
```

## Steps

1. Enable post-build inspect for merge verification leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Output = "sandbox.bin"
	req.OutputSet = true
	req.AfterBuildInspect = true
	req.BuildTwice = false
	return nil
}
```
