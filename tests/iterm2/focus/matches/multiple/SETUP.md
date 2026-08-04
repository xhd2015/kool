# Scenario

**Feature**: ambiguous candidates require explicit selection

```text
focus target -> FocusFake candidates [0, 1] -> error or --index selection
```

```go
import (
 iterm2cmd "github.com/xhd2015/kool/tools/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
 target := d.DOCTEST_CASE
 req.Candidates = []iterm2cmd.FocusCandidate{
  {WindowID: "1", WindowTitle: "credit-pricing", TabIndex: 2, SessionID: "w0t2p0", Path: target},
  {WindowID: "3", WindowTitle: "worktrees", TabIndex: 1, SessionID: "w2t1p0", KoolTargetDir: target},
 }
 return nil
}
```
