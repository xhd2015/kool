# Scenario

```text
session send -h -> documents --tab / --tab-index / --session-id
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"send", "-h"}
	return nil
}
```
