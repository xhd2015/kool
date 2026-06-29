# Scenario

**Feature**: CLI surfaces install, osascript, and platform errors

```
kool iterm2 -> OpenConfig failure -> stderr + non-zero exit
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Phase == "" {
		req.Phase = "cli"
	}
	if req.InstalledEnv == "" {
		req.InstalledEnv = "1"
	}
	return nil
}
```