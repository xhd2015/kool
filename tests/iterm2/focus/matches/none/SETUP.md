# Scenario

**Feature**: no matching iTerm2 session

```text
focus target -> FocusFake [] -> fatal no-match error
```

```go
import iterm2cmd "github.com/xhd2015/kool/tools/iterm2"

func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Candidates = []iterm2cmd.FocusCandidate{}; return nil }
```
