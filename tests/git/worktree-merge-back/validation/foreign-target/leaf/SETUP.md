# Scenario

**Feature**: invoke merge-back with foreign --to target

```
user -> merge-back --to <foreign-wt> -> validation error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.ForeignWT == "" || req.To != req.ForeignWT {
		t.Fatal("expected foreign target from ancestor setup")
	}
	return nil
}
```