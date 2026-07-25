# Scenario

**Feature**: help mentions --no-enrich and --no-tree (or tree-related flag)

```
kool iterm2 sessions -h
  -> stdout contains --no-enrich
  -> stdout contains --no-tree (or "tree" flag wording covering tree suppress)
```

## Steps

1. Default help args from parent (sessions -h).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// ModeHelp + HelpArgs set by parent help/SETUP.md
	return nil
}
```
