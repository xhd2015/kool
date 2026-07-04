#!/usr/bin/env bash
set -euo pipefail

# Build __PROJECT_NAME__.app and install to /Applications (no drag-and-drop).
# Same end result as a standard .app install: copy into /Applications and clear
# the quarantine extended attribute.
#
# Override via env (used by install-debug.sh):
#   APP_NAME, BUNDLE_ID, SWIFT_BUILD_CONFIG, INSTALL_VARIANT

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

APP_NAME="${APP_NAME:-__PROJECT_NAME__}"
BUNDLE_ID="${BUNDLE_ID:-__BUNDLE_ID__}"
SWIFT_BUILD_CONFIG="${SWIFT_BUILD_CONFIG:-release}"
INSTALL_VARIANT="${INSTALL_VARIANT:-release}"
SOURCE_APP="$PROJECT_DIR/$APP_NAME.app"
INSTALL_ROOT="${INSTALL_ROOT:-/Applications}"
TARGET_APP="$INSTALL_ROOT/$APP_NAME.app"
OPEN_AFTER_INSTALL=1

usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Build $APP_NAME.app (via bundle.sh) and install to $INSTALL_ROOT.

Options:
  --no-open       Skip launching $APP_NAME after install
  --open          Launch $APP_NAME after install (default)
  --install-root  Override install directory (default: /Applications)
  -h, --help      Show this help

Examples:
  ./script/install.sh
  ./script/install.sh --no-open
  INSTALL_ROOT=\$HOME/Applications ./script/install.sh
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --no-open)
            OPEN_AFTER_INSTALL=0
            shift
            ;;
        --open)
            OPEN_AFTER_INSTALL=1
            shift
            ;;
        --install-root)
            INSTALL_ROOT="$2"
            TARGET_APP="$INSTALL_ROOT/$APP_NAME.app"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [[ ! -w "$INSTALL_ROOT" ]]; then
    echo "error: cannot write to $INSTALL_ROOT (try sudo or set INSTALL_ROOT)" >&2
    exit 1
fi

echo "==> Building $APP_NAME.app ($SWIFT_BUILD_CONFIG)"
BUNDLE_SKIP_DMG=1 \
    APP_NAME="$APP_NAME" \
    BUNDLE_ID="$BUNDLE_ID" \
    SWIFT_BUILD_CONFIG="$SWIFT_BUILD_CONFIG" \
    "$SCRIPT_DIR/bundle.sh"

if [[ ! -d "$SOURCE_APP" ]]; then
    echo "error: expected app bundle at $SOURCE_APP" >&2
    exit 1
fi

echo "==> Stopping running $APP_NAME (if any)"
osascript -e "tell application \"$APP_NAME\" to quit" 2>/dev/null || true
pkill -f "${TARGET_APP}/Contents/MacOS/" 2>/dev/null || true
sleep 0.5

echo "==> Installing to $TARGET_APP"
rm -rf "$TARGET_APP"
ditto "$SOURCE_APP" "$TARGET_APP"
xattr -dr com.apple.quarantine "$TARGET_APP" 2>/dev/null || true

echo ""
echo "==> Installed: $TARGET_APP"
if [[ "$INSTALL_VARIANT" == "debug" ]]; then
    echo "    Bundle ID: $BUNDLE_ID"
    echo "    Daemon state: ~/__STATE_SUBPATH__-debug (isolated from production)"
    echo "    Daemon port: random available port on each launch (not __DEFAULT_PORT__)"
else
    echo "    First launch from a downloaded build may still need right-click → Open."
    echo "    Local ad-hoc builds from this machine usually launch normally."
    echo "    Daemon port: __DEFAULT_PORT__"
fi

if [[ "$OPEN_AFTER_INSTALL" -eq 1 ]]; then
    echo "==> Opening $APP_NAME"
    open "$TARGET_APP"
fi