# Scenario

**Feature**: `--command` patches only the command field

```
update bots --tab-id a --command "echo A-new"
  -> tab a command becomes echo A-new; name/cwd/no_submit/window unchanged
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=a; Command=`echo A-new`.

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
	req.Command = "echo A-new"
	return nil
}
```
