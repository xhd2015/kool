# Scenario

**Feature**: show unknown set name errors

```
show no-such-set -> Error, exit ≠ 0
```

## Steps

1. SetName that has no JSON file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.SetName = "no-such-set"
	return nil
}
```
