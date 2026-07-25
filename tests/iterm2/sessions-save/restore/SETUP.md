# Scenario

**Feature**: sessions restore — recreate windows from checkpoint

```
Caller
  -> sessions restore [--dry-run] [--file] [--color|--no-color]
  -> read SaveDocument
  <- plan / apply / consumed error
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
