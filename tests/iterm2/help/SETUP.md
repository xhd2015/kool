# Scenario

**Feature**: CLI help output

```
kool iterm2 --help -> usage on stdout -> exit 0
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "cli"
	return nil
}
```