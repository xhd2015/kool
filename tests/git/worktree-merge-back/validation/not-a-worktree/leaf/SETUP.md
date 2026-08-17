# Scenario

**Feature**: invoke merge-back from main repo cwd

```
user (cwd=main) -> merge-back handler -> validation error
```

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.MainRepo == "" {
		t.Fatal("expected MainRepo from ancestor setup")
	}
	return nil
}
```