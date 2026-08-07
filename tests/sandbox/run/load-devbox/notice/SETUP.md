# Scenario

**Feature**: successful load emits notice on stdout

```
./primary --load-devbox ABS -- sh -c 'true'
  -> RunStdout contains "notice: loading devbox <abs>"; no ANSI in capture
```

## Steps

1. Notice leaves focus on stdout notice contract (non-TTY capture → no grey ANSI).
