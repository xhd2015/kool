# Scenario

**Feature**: `-r` miss path — new window and cd in generated script

```
kool iterm2 -r <dir> -> ModeReuseCurrent script -> else branch: create window + cd
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Send = nil
	return nil
}
```