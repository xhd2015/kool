# Scenario

**Feature**: --list omits a nested separate git repo's module line

```
# ext/ has its OWN .git from a separate `git init ext`; root never tracked it as a submodule
root + ext/go.mod + ext/.git(separate repo, untracked by root) -> stdout:
  . some.com/root
  (no ext line)
```

## Steps

1. Root workspace (`some.com/root`, git-init'd) is created by the `list/` grouping Setup
   (root committed before `ext/` exists, so `ext/` is never in the root index).
2. Create `ext/go.mod` (`some.com/ext`) and init `ext/` as its **own** git repo (separate
   `.git` inside `ext/`, NOT a submodule of root).

```go
func Setup(t *testing.T, req *Request) error {
	ext := filepath.Join(req.RootDir, "ext")
	writeModule(t, ext, "some.com/ext")
	initGitRepo(t, ext) // separate .git inside ext/, NOT a submodule of root
	return nil
}
```
