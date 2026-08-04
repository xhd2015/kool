# Scenario

**Feature**: one `user.koolTargetDir` match focuses

```text
focus target -> FocusFake candidate user.koolTargetDir=target -> focus session
```

```go
import (
 iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
 req.Candidates = []iterm2cmd.FocusCandidate{{WindowID: "3", WindowTitle: "worktrees", TabIndex: 1, SessionID: "w2t1p0", KoolTargetDir: d.DOCTEST_CASE}}
 return nil
}
```
