# Scenario

**Feature**: install flag validation (hard errors before download)

```
# incompatible flags
user -> kool iterm2 install --via-open --download-only …
  -> Error: on stderr; non-zero exit; no zip download
```

## Steps

1. Children set conflicting flags; optional fake HTTP only if resolve might run
   (this suite injects HTTP so a late resolve cannot hit the real network).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Help = false
	// Inject fake HTTP so a buggy late resolve cannot call iterm2.com.
	req.UseFakeHTTP = true
	return nil
}
```
