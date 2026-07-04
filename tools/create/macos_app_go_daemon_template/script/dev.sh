#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SWIFT_DIR="$PROJECT_DIR/__PROJECT_NAME__-swift"

echo "==> Building __DAEMON_NAME__"
mkdir -p "$PROJECT_DIR/.build"
cd "$PROJECT_DIR/go-pkgs/cmd/__DAEMON_NAME__"
go build -o "$PROJECT_DIR/.build/__DAEMON_NAME__" .

echo "==> Building __PROJECT_NAME__ in $SWIFT_DIR"
cd "$SWIFT_DIR"
swift build

echo "==> Starting __PROJECT_NAME__"
export DAEMON_CLI="$PROJECT_DIR/.build/__DAEMON_NAME__"
swift run __PROJECT_NAME__