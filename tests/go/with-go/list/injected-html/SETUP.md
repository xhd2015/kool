# Scenario

**Feature**: injected FetchHTML matching-divs become `go…` stdout lines

```
FetchHTML(two id="go…" divs) -> ListWith -> "go1.22.1\ngo1.19.13\n"
```

## Preconditions

- HTML includes two version divs plus noise (`<div>` without id, and
  `id="something-else"`). Only `id="go…"` on a `<div ` line counts
  (same parse as xgo `support/downloadgo` tests/list/matching-divs).

## Steps

1. Inject `FetchHTML` that returns the fixture HTML below.

```go
import (
	"context"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const matchingHTML = `
<html>
<body>
<div class="toggle" id="go1.22.1"></div>
<div id="go1.19.13"></div>
<div>no version</div>
<div id="something-else">skip</div>
</body>
</html>
`

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchHTML = func(ctx context.Context) (string, error) {
		return matchingHTML, nil
	}
	return nil
}
```
