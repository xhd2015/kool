# Scenario

**Feature**: list shows a configured set name

```
bots.json in config dir -> list -> stdout contains bots and tab count hint
```

## Steps

1. Write `bots.json` fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeBotsConfig(t, req.ConfigDir)
	return nil
}
```
