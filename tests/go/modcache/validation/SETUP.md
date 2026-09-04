# Scenario

**Feature**: modcache routing and path validation fail fast

```
user -> modcache [no sub | nosuch | --modcache missing]
  -> non-zero; Error: or usage on stderr
```
