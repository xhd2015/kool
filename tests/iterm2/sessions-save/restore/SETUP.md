# Scenario

**Feature**: sessions restore — recreate windows from checkpoint (+ Space placement)

```
Caller
  -> sessions restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space]
  -> read SaveDocument; plan space N (Desktop N+1) unless ignore
  <- plan / apply (Switch+Create+AS) / consumed error
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
