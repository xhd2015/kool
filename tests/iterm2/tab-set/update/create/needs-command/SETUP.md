# Scenario

**Feature**: `--create` without `--command` is an error

```
update bots --tab-id c --create
  -> error (command required); bots.json unchanged
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=c; Create=true; Command empty.

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
	// Create inherited true; Command left empty.
	return nil
}
```
