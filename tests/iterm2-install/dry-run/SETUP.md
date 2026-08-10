# Scenario

**Feature**: install dry-run with fake HTTP latest (no zip write)

```
# dry-run resolves URL then prints plan
user -> kool iterm2 install --dry-run [--via-open] --download-dir DIR
  -> fake HTTP latest -> plan on stdout; exit 0; no zip under DIR
```

## Steps

1. Enable DryRun + UseFakeHTTP for all children.
2. Leaves set ViaOpen true/false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DryRun = true
	req.UseFakeHTTP = true
	req.Help = false
	return nil
}
```
