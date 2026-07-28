# Scenario

**Feature**: `--rm` is exclusive with patch flags

```
update bots --tab-id a --rm --force --no-submit
  -> exclusive error; file unchanged
```

## Steps

1. Write bots fixture.
2. Rm=true; Force=true; UpdateNoSubmit=true (patch flag conflict).

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
	req.UpdateNoSubmit = true // exclusive with --rm
	return nil
}
```
