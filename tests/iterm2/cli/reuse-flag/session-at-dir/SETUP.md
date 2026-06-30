# Scenario

**Feature**: `-r` match path — focus existing session/tab only

```
kool iterm2 -r <dir> -> scan finds path == targetDir -> focus session/tab (no cd, no tab create)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Send = nil
	return nil
}
```