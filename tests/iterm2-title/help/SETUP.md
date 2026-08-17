# Scenario

**Feature**: kool iterm2 help documents title subcommands

```
# help path
kool iterm2 --help
  -> usage on stdout lists set-title and get-title
```

## Steps

1. Enable Help flag for descendants.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Help = true
	return nil
}
```
