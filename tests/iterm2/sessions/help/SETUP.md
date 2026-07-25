# Scenario

**Feature**: sessions help surface (no capture)

```
# user requests sessions help
user -> kool iterm2 sessions -h|--help
  -> usage on stdout; exit 0; no iTerm query
```

## Steps

1. Set Mode=help; default HelpArgs to sessions -h.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeHelp
	if len(req.HelpArgs) == 0 {
		req.HelpArgs = []string{"sessions", "-h"}
	}
	return nil
}
```
