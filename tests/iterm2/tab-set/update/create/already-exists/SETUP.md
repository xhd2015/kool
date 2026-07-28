# Scenario

**Feature**: `--create` on an existing tab id is an error

```
update bots --tab-id a --create --command "echo other"
  -> error (already exists); bots.json unchanged
```

## Steps

1. Write bots fixture (id a exists).
2. SetName=bots; TabID=a; Create=true; Command=`echo other`.

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
	req.Command = "echo other"
	return nil
}
```
