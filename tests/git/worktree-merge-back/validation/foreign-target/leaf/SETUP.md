# Scenario

**Feature**: invoke merge-back with foreign --to target

```
user -> merge-back --to <foreign-wt> -> validation error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.ForeignWT == "" || req.To != req.ForeignWT {
		t.Fatal("expected foreign target from ancestor setup")
	}
	return nil
}
```