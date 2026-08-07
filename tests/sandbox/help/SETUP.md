# Scenario

**Feature**: sandbox help mode (no build)

```
# user asks for help at root or build
user -> kool sandbox [--help | build --help]
  -> handler prints usage, exit 0 (no seal, no -o)
```

## Steps

1. Mark help branch; no Subcommand build path unless leaf sets HelpBuild.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Help leaves finish immediately.
	if req.ProcessTimeout > 15*time.Second || req.ProcessTimeout <= 0 {
		req.ProcessTimeout = 15 * time.Second
	}
	req.Subcommand = ""
	req.AfterBuildInspect = false
	req.BuildTwice = false
	return nil
}
```
