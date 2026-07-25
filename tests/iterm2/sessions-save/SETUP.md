# Scenario

**Feature**: nested doctest root for `kool iterm2 sessions save|restore`

```
# no live iTerm; fixtures via InstallPhasedFixtureCollectorForTest
Caller
  -> kool iterm2 sessions save|restore [--dry-run] [--file] [--color|--no-color]
  -> SnapshotCollector (phased fixture)
  -> critical filter (grok/codex/mark)
  -> plan stream / file write / restore plan
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
