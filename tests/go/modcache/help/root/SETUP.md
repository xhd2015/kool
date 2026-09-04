# Scenario

**Feature**: root --help lists inspect and prune

```
user -> HandleWith(["--help"])
  -> stdout usage; exit 0
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpAtRoot = true
	return nil
}
```
