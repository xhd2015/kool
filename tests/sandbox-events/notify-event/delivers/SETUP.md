# Scenario

**Feature**: notify-event dials live socks and delivers JSON event

```
# mock unix listener on ROOT/events/sess-mock.sock
kool sandbox notify-event --type devbox.updated --path ABS --root ROOT
  -> mock receives {"v":1,"type":"devbox.updated","path":ABS,"ts":...}
```

## Steps

1. MockListener=true; RunNotifyEvent; assert DeliveredRaw / FirstDelivered.
