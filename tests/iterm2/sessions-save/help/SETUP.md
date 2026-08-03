# Scenario

**Feature**: help text for sessions save / restore

```
Caller
  -> kool iterm2 sessions [-h|save -h|restore -h]
  <- usage documents subcommands, color flags, --ignore-macos-space
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
