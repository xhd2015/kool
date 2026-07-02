# Scenario

**Feature**: new-window mode with --send commands

```
# -n --send → ModeForceNew + follow-up commands in new window
kool iterm2 <dir> -n --send "echo hi" -> ModeForceNew -> new window + write text "echo hi"
```

## Preconditions

- Args = ["-n", "--send", "echo hi"]

## Steps

1. Run with -n and --send flags

## Context

- --send should generate `write text` commands in the script

```go
func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"-n", "--send", "echo hi"}
    return nil
}
```
