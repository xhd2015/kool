# Scenario

**Feature**: `--tab next` sends to the next tab’s iTerm session

```text
current=tab1 → session send --tab next "from-next" → SendText(tab2)
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CurrentSessionID = fixtureTab1ID
	req.Args = []string{"send", "--tab", "next", "from-next"}
	return nil
}
```
