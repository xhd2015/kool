# Scenario

**Feature**: `sessions -h` lists auto-backup subcommand

```
Caller
  -> sessions -h
  <- mentions auto-backup (alongside save / restore / snapshot when present)
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
