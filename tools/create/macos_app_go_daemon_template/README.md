# __PROJECT_NAME__

macOS menu bar app with a Go HTTP daemon.

## Requirements

- macOS 13.0+
- Swift 5.9+ (Xcode 15+ or standalone toolchain)
- Go 1.23+

## Layout

```
__PROJECT_NAME__/
├── __PROJECT_NAME__-swift/   # Swift menu bar app (Package.swift + sources)
├── go-pkgs/                  # Go HTTP daemon
└── script/
```

## Quick Start

```sh
./script/dev.sh
```

## Build app bundle

```sh
./script/bundle.sh
```

## Install to /Applications

```sh
./script/install.sh
```

Debug variant (isolated bundle ID and state dir):

```sh
./script/install-debug.sh
```

## Daemon

The app spawns `__DAEMON_NAME__ serve` on `127.0.0.1:__DEFAULT_PORT__`.

State directory: `$HOME/__STATE_SUBPATH__`

```sh
# Manual daemon (optional)
go build -o .build/__DAEMON_NAME__ ./go-pkgs/cmd/__DAEMON_NAME__
.build/__DAEMON_NAME__ serve
.build/__DAEMON_NAME__ status
```