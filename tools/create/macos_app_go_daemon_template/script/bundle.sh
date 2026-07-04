#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SWIFT_DIR="$PROJECT_DIR/__PROJECT_NAME__-swift"

APP_NAME="${APP_NAME:-__PROJECT_NAME__}"
BUNDLE_ID="${BUNDLE_ID:-__BUNDLE_ID__}"
SWIFT_BUILD_CONFIG="${SWIFT_BUILD_CONFIG:-release}"
SWIFT_EXECUTABLE="${SWIFT_EXECUTABLE:-__PROJECT_NAME__}"
BUNDLE_DIR="$PROJECT_DIR/$APP_NAME.app"
CONTENTS="$BUNDLE_DIR/Contents"
MACOS_BIN="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"

echo "==> Building __DAEMON_NAME__ CLI"
mkdir -p "$PROJECT_DIR/.build"
cd "$PROJECT_DIR/go-pkgs/cmd/__DAEMON_NAME__"
go build -o "$PROJECT_DIR/.build/__DAEMON_NAME__" .

echo "==> Building $APP_NAME ($SWIFT_BUILD_CONFIG, bundle: $BUNDLE_ID)"
cd "$SWIFT_DIR"
swift build -c "$SWIFT_BUILD_CONFIG"

echo "==> Creating .app bundle at $BUNDLE_DIR"
rm -rf "$BUNDLE_DIR"
mkdir -p "$MACOS_BIN" "$RESOURCES"

BIN_PATH="$(swift build -c "$SWIFT_BUILD_CONFIG" --show-bin-path)/$SWIFT_EXECUTABLE"
cp "$BIN_PATH" "$MACOS_BIN/$SWIFT_EXECUTABLE"
cp "$PROJECT_DIR/.build/__DAEMON_NAME__" "$MACOS_BIN/__DAEMON_NAME__"
chmod +x "$MACOS_BIN/__DAEMON_NAME__"

cat > "$CONTENTS/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$SWIFT_EXECUTABLE</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>13.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

echo "==> Ad-hoc code signing"
codesign --force --deep -s - "$BUNDLE_DIR" 2>/dev/null || true

echo ""
echo "==> App bundle ready: $BUNDLE_DIR"

if [[ "${BUNDLE_SKIP_DMG:-}" == "1" ]]; then
    echo "    (DMG skipped — set BUNDLE_SKIP_DMG=0 or unset to create $APP_NAME.dmg)"
    exit 0
fi

DMG_PATH="$PROJECT_DIR/$APP_NAME.dmg"
STAGING="$PROJECT_DIR/.dmg-staging"

echo "==> Creating DMG at $DMG_PATH"
rm -rf "$STAGING" "$DMG_PATH"
mkdir -p "$STAGING"

cp -R "$BUNDLE_DIR" "$STAGING/"
ln -s /Applications "$STAGING/Applications"

hdiutil create -volname "$APP_NAME" \
    -srcfolder "$STAGING" \
    -ov -format UDZO \
    "$DMG_PATH"

rm -rf "$STAGING"

echo ""
echo "==> Done:"
echo "    DMG:  $DMG_PATH"
echo ""
echo "    To install on another machine:"
echo "      1. Copy $APP_NAME.dmg to target machine"
echo "      2. Open the DMG, drag $APP_NAME.app to the Applications folder"
echo "      3. First launch: right-click → Open (Gatekeeper bypass)"