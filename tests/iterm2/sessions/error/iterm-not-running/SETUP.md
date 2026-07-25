# Scenario

**Feature**: snapshot fails when iTerm2 is not running

```
InstallPhasedFixtureCollectorForTest(ITermRunning=false)
  -> sessions snapshot --no-color
  -> exit non-zero; stderr mentions iTerm / not running
```

## Steps

1. ITermRunning=false; still install fixture for inject flag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.ITermRunning = boolPtr(false)
	req.UseTwoWindowFixture = true
	req.NoColor = true
	return nil
}
```
