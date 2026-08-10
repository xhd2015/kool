# Scenario

**Feature**: hot-reload of runtime-load file layer on devbox.updated

```
# live session with load seal L containing reload-me.txt=old
# rebuild L with reload-me.txt=new; notify-event --path L
  -> session root reload-me.txt becomes new without guest restart
```

## Steps

1. Secondary load pack; primary RuntimeLoadDevbox; notify after rebuild.
