# Scenario

**Feature**: garbage duration string is rejected

```
kool for-every notaduration --max-runs 1 true
  -> non-zero; stderr mentions duration / invalid
```

## Steps

1. Set Duration to a non-parseable token.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Duration = "notaduration"
	return nil
}
```
