# Scenario

**Feature**: kool iterm2 new-window flag

```
# user specifies mode via flags
kool iterm2 <dir> [-n | --new-window | -r | --reuse | --reuse-window] -> CLI handler -> lib.OpenConfig(cfg)
# library selects script builder based on Mode
cfg.Mode -> {ModeSmart, ModeReuseCurrent, ModeForceNew} -> BuildScript
# osascript runs the generated AppleScript
BuildScript -> osascript -> iTerm2
```

## Preconditions

- `KOOL_ITERM2_GOOS=darwin` and `KOOL_ITERM2_INSTALLED=1` env vars are set by root Run
- A temp dir is created for the test, used as both target dir and working dir
- `KOOL_ITERM2_SCRIPT_OUT` captures the generated AppleScript to a file

## Steps

1. Build full args from Request.Args + temp dir
2. Invoke `iterm2cmd.RunForTest` with env overrides
3. Read captured script from `KOOL_ITERM2_SCRIPT_OUT` path

## Context

- Temp dir is created via `t.TempDir()` — automatically cleaned up
- Library and handler are from the local workspace (replace directives in go.mod)
