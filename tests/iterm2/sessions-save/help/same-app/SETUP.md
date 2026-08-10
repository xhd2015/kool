# Scenario

**Feature**: restore help documents `--same-app` and default prefer-home when multiple installs

```
Caller
  -> sessions restore -h
  <- exit 0; documents --same-app + prefer ~/Applications when multiple installs
```

## Steps

1. ModeHelp with restore -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	req.HelpArgs = []string{"sessions", "restore", "-h"}
	return nil
}
```
