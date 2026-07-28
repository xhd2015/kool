# Scenario

**Feature**: `--create --command` appends a new tab id

```
update bots --tab-id c --create --command "echo c"
  -> bots.json has tabs a, b, c; c.command=echo c
```

## Steps

1. Write bots fixture (ids a,b only).
2. SetName=bots; TabID=c; Create=true (inherited); Command=`echo c`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.TabID = "c"
	req.Command = "echo c"
	return nil
}
```
