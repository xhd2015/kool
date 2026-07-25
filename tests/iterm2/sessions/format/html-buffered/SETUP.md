# Scenario

**Feature**: --html emits one full buffered HTML document

```
sessions snapshot --html --no-color
  -> stdout contains <html
  -> fixture session present
```

## Steps

1. HTML=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.HTML = true
	return nil
}
```
