# Scenario

**Feature**: list shows multi-task JSONC fixture sorted by label

```
tasks.json with // comments, trailing commas, shell/process/composite
  -> list: Build All, Compile, Serve; types and BG hints
```

## Steps

1. Write multi-task fixture; Dir=WorkingDir.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeMultiTaskFixture(t, req.WorkingDir)
	req.Dir = req.WorkingDir
	return nil
}
```
