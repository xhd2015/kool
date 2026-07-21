# Scenario

**Feature**: CLI surfaces install, osascript, and platform errors

```
kool iterm2 -> OpenConfig failure -> stderr + non-zero exit
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	if req.Phase == "" {
		req.Phase = "cli"
	}
	if req.InstalledEnv == "" {
		req.InstalledEnv = "1"
	}
	return nil
}

// markErrorTree keeps hierarchical child packages importing this package live.
func markErrorTree() {}
```