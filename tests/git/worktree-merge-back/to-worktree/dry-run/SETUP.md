# Scenario

**Feature**: dry-run merge-back with --to sibling

```
user -> merge-back --to sibling --dry-run -> planned commands only
```

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.Remove = false
	req.ConfirmFromStdin = false
	if req.SiblingPath == "" || req.To != req.SiblingPath {
		t.Fatal("expected sibling target from ancestor setup")
	}
	return nil
}
```