## Steps
- Set req.Dir to a non-existent directory path
- Set RefA and RefB to arbitrary values

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Dir = filepath.Join(req.Dir, "nonexistent")
	req.RefA = "main"
	req.RefB = "main"
	return nil
}
```
