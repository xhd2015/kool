# Scenario

**Feature**: `--name` and `--cwd` patch display name and working directory

```
update bots --tab-id a --name "Alpha" --cwd /var/tmp
  -> tab a name=Alpha, cwd=/var/tmp; command unchanged
```

## Steps

1. Write bots fixture.
2. SetName=bots; TabID=a; TabName=Alpha; Cwd=/var/tmp.

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
	req.TabName = "Alpha"
	req.Cwd = "/var/tmp"
	return nil
}
```
