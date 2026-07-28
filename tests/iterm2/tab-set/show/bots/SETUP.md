# Scenario

**Feature**: show bots fixture details

```
show bots -> local-bots, tab ids a/b, commands echo a / echo b
```

## Steps

1. Write bots.json; SetName=bots.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	return nil
}
```
