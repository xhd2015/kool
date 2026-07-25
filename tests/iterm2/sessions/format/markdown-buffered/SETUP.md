# Scenario

**Feature**: --markdown emits one full buffered Markdown document

```
sessions snapshot --markdown --no-color
  -> stdout starts with # iTerm2 snapshot (or contains that heading)
  -> includes fixture session material
```

## Steps

1. Markdown=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Markdown = true
	return nil
}
```
