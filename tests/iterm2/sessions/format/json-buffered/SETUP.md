# Scenario

**Feature**: --json emits one valid JSON document (buffered)

```
sessions snapshot --json --no-color
  -> stdout is a single json.Unmarshal object with windows array
  -> not progressive multi-document stream
```

## Steps

1. JSON=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	return nil
}
```
