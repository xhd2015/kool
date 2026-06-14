## Steps
- Set RefA and RefB to arbitrary values — neither will resolve since there is no git repo

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RefA = "main"
	req.RefB = "main"
	return nil
}
```
