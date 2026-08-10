# Scenario

**Feature**: save help documents multi-app / app field / preferred restore

```
Caller
  -> sessions save -h
  <- exit 0; mentions dual installs or app paths or preferred restore
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
