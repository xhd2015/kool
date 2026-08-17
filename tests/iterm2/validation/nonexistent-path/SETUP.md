# Scenario

**Feature**: nonexistent directory path

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DirPath = filepath.Join(req.WorkingDir, "no-such-dir")
	return nil
}
```