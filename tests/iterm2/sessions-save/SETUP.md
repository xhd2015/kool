# Scenario

**Feature**: nested doctest root for `kool iterm2 sessions save|restore` (+ macOS Space)

```
# no live iTerm / Mission Control; fixtures via InstallPhasedFixtureCollectorForTest
Caller
  -> kool iterm2 sessions save|restore [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space]
  -> SnapshotCollector (phased fixture) + optional SpaceIndexForWindow inject
  -> critical filter (grok/codex/mark) + space record / placement plan
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
