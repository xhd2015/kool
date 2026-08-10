# Scenario

**Feature**: ensure missing bin runs InstallCmd once

```
LookPath miss
  -> kool codex install
  -> exit 0
  -> ShellCalls == [install.InstallCmd]
  -> FetchLatestCalls == 0
```

## Steps

1. Present=false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Present = false
	req.LocalRaw = ""
	req.Latest = ""
	return nil
}
```
