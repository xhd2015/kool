# Scenario

**Feature**: save help documents `--spaces`

```
Caller
  -> sessions save -h
  <- exit 0; mentions --spaces
```

## Steps

1. ModeHelp with save -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	req.HelpArgs = []string{"sessions", "save", "-h"}
	return nil
}
```
