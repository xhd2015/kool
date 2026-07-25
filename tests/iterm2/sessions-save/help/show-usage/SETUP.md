# Scenario

**Feature**: `kool iterm2 sessions -h` documents save and restore

```
Caller
  -> sessions -h
  <- mentions save, restore, snapshot
```

## Steps

1. ModeHelp; HelpArgs = sessions -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	req.HelpArgs = []string{"sessions", "-h"}
	return nil
}
```
