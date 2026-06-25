# Scenario

**Feature**: invoke merge-back from main repo cwd

```
user (cwd=main) -> merge-back handler -> validation error
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.MainRepo == "" {
		t.Fatal("expected MainRepo from ancestor setup")
	}
	return nil
}
```