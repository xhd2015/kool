# Scenario

**Feature**: `--dry-run --via-open` plan mentions open / via-open mode

```
kool iterm2 install --dry-run --via-open --download-dir DIR
  + fake HTTP latest
  -> exit 0
  -> stdout dry-run banner + version 3.6.11
  -> steps/mode mention open or clear-quarantine or via-open
  -> no zip written under DIR
```

## Steps

1. Set ViaOpen=true (DryRun + UseFakeHTTP inherited).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ViaOpen = true
	return nil
}
```
