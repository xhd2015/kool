# Scenario

**Feature**: sessions help documents P4 enrich flags

```
# user requests sessions help
user -> kool iterm2 sessions -h|--help
  -> usage lists --no-enrich and tree-related flag; exit 0; no capture
```

## Steps

1. Mode=help; default HelpArgs to sessions -h.

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
