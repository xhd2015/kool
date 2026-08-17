# Scenario

**Feature**: path exists but is not a directory

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	file := filepath.Join(req.WorkingDir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		return err
	}
	req.DirPath = file
	return nil
}
```