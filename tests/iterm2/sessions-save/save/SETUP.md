# Scenario

**Feature**: sessions save — checkpoint critical tabs (+ Space + multi-app `app`)

```
Caller
  -> sessions save [--dry-run] [--file] [--color|--no-color] [--ignore-macos-space] [--spaces]
  -> preflight app + multi-app merge + critical filter + SpaceIndexForWindow (unless ignore)
  -> stream window plan / write JSON / error
```
