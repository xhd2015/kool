# Scenario

**Feature**: restore skips checkpoint tabs whose same grok/codex session or mark
message is already live in iTerm2 (warn + skip; dry-run full non-mutating plan)

```
Caller
  -> seed checkpoint (critical tabs)
  -> install live phased fixture (matching / miss / fail capture)
  -> sessions restore [--dry-run]
  <- warn on hit; skip already-running; remaining would restore / AS; stamp rules
```

## Steps

1. ModeRestore (shared by already-running leaves). Leaves set DryRun / live
   fixture / MockRestoreAS / SeedDoc.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeRestore
	return nil
}
```
