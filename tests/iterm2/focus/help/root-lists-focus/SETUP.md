# Scenario

**Feature**: iTerm2 parent help names focus

```text
kool iterm2 --help -> parent usage includes focus
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error { req.Args = []string{"--help"}; return nil }
```
