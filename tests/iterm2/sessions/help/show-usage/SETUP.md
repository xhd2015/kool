# Scenario

**Feature**: sessions help documents snapshot and P3 stream flag

```
kool iterm2 sessions -h
  -> stdout mentions snapshot, --json, --no-stream; exit 0; trailing newline
```

## Steps

1. Use default sessions -h HelpArgs from parent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Explicit help args for this leaf (sessions root help).
	req.HelpArgs = []string{"sessions", "-h"}
	return nil
}
```
