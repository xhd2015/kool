# Scenario

**Feature**: bare modcache without subcommand is an error

```
user -> HandleWith([])
  -> non-zero; stderr hints inspect/prune/help
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{}
	return nil
}
```
