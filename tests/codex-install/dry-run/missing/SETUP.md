# Scenario

**Feature**: dry-run with missing codex plans install (InstallCmd); no shell

```
LookPath miss
  -> kool codex install --dry-run
  -> exit 0
  -> stdout dry-run + would install / install plan
  -> stdout mentions InstallCmd (or install script curl)
  -> ShellCalls empty
  -> FetchLatestCalls == 0
```

## Steps

1. Present=false (LookPath miss).

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
