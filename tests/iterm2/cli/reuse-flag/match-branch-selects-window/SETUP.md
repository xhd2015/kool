# Scenario

**Feature**: `-r` match branch brings matchingWindow to front before focusing tab/session

```
kool iterm2 -r <dir> -> match branch: select matchingWindow -> select tab/session (not background window only)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Send = nil
	return nil
}
```