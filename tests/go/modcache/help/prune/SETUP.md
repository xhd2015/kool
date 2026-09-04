# Scenario

**Feature**: prune --help documents --dry-run

```
user -> HandleWith(["prune", "--help"])
  -> usage on stdout; exit 0
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HelpPrune = true
	return nil
}
```
