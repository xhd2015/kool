# Scenario

**Feature**: `--rm --force` removes one tab; others remain

```
update bots --tab-id a --rm --force
  -> tab a gone; tab b remains; exit 0
```

## Steps

1. Write bots fixture (a,b).
2. SetName=bots; TabID=a; Rm=true (inherited); Force=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.TabID = "a"
	req.Force = true
	return nil
}
```
