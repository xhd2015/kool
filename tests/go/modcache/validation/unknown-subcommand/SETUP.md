# Scenario

**Feature**: unknown subcommand is an error

```
user -> HandleWith(["nosuch"])
  -> non-zero; Error: on stderr
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"nosuch"}
	return nil
}
```
