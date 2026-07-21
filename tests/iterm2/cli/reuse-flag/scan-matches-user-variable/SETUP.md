# Scenario

**Feature**: `-r` scan matches both session `path` and `user.koolTargetDir`

Second `kool iterm2 -r <dir>` should focus the window opened by the first run even
when iTerm's `path` variable has not updated yet. The scan loop must treat
`user.koolTargetDir` as an alternate match key alongside `path`.

```
kool iterm2 -r <dir> -> scan: path == targetDir OR user.koolTargetDir == targetDir
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markCliReuseFlagTree()
	markCliTree()
	markRootTree()
	req.Send = nil
	return nil
}
```