# Scenario

**Feature**: dry-run without `--via-open` does not claim via-open mode

```
kool iterm2 install --dry-run --download-dir DIR
  + fake HTTP latest
  -> exit 0; dry-run banner + version
  -> stdout must NOT claim via-open / user-open mode
  -> no zip written
```

## Steps

1. ViaOpen remains false (default).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ViaOpen = false
	return nil
}
```
