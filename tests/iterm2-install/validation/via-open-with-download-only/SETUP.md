# Scenario

**Feature**: `--via-open` + `--download-only` is a hard error

```
kool iterm2 install --via-open --download-only --download-dir DIR
  -> non-zero exit
  -> stderr contains Error:
  -> no zip under DIR
```

## Steps

1. Set both ViaOpen and DownloadOnly (mutually exclusive).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ViaOpen = true
	req.DownloadOnly = true
	req.DryRun = false
	return nil
}
```
