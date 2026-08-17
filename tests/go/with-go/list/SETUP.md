# Scenario

**Feature**: ListWith prints go-prefixed versions from injected FetchHTML

```
# no network; downloadgo.List parses id="go…" divs; ListWith writes go%s\n
FetchHTML(HTML) -> ListWith -> stdout go1.22.1 / go1.19.13
```

## Steps

1. Set `req.Op` to `list` and record FetchHTML calls.
2. Leaf Setup injects HTML.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "list"
	req.RecordFetch = true
	return nil
}
```
