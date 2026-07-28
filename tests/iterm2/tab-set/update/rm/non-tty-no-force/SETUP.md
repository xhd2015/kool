# Scenario

**Feature**: non-TTY `--rm` without `--force` errors; file unchanged

```
update bots --tab-id a --rm   (handler non-TTY, no --force)
  -> exit ≠ 0; bots.json still has tabs a,b
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=a; Rm=true; Force=false.

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
	req.Force = false
	return nil
}
```
