# Scenario

**Feature**: dest `$InstallDir/go1.19.13` already exists as a directory

```
# fixture dest dir present before ResolveGorootWith
InstallDir + go1.19.13 dir -> ResolveGorootWith -> dest path; Install unused
```

## Steps

1. Create `$InstallDir/go1.19.13` as a directory.
2. Leaf Setup sets version spelling and Prompt.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if err := os.MkdirAll(destPin(req.InstallDir), 0755); err != nil {
		return err
	}
	return nil
}
```
