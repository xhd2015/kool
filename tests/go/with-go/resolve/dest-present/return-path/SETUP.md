# Scenario

**Feature**: an existing dest directory is returned; Install and Prompt are unused

```
# $InstallDir/go1.19.13 already a directory
dest dir present -> ResolveGorootWith(go1.19) -> dest; Install not called; Prompt not written
```

## Steps

1. Keep `GoVersion` `go1.19`.
2. Set a Prompt that must not appear on Stderr (install will not run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.GoVersion = "go1.19"
	req.Prompt = "should-not-print\n"
	return nil
}
```
