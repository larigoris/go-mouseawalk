#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="${1:-GOwalk}"
OUTPUT_DIR="${2:-$ROOT_DIR/dist}"
TARGET_ARCH="${GOARCH:-$(uname -m)}"
ICONS_DIR="$ROOT_DIR/icons"

case "$TARGET_ARCH" in
  arm64)
    GOARCH_VALUE="arm64"
    ;;
  x86_64)
    GOARCH_VALUE="amd64"
    ;;
  *)
    echo "Unsupported architecture: $TARGET_ARCH" >&2
    exit 1
    ;;
esac

BUNDLE_PATH="$OUTPUT_DIR/$APP_NAME.app"
MACOS_DIR="$BUNDLE_PATH/Contents/MacOS"
RESOURCES_DIR="$BUNDLE_PATH/Contents/Resources"
PLIST_PATH="$BUNDLE_PATH/Contents/Info.plist"

rm -rf "$BUNDLE_PATH"
mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

cd "$ROOT_DIR"

ICON_FILE=""
if [[ -f "$ICONS_DIR/icon.icns" ]]; then
  ICON_FILE="$ICONS_DIR/icon.icns"
elif [[ -d "$ICONS_DIR/icon.iconset" ]]; then
  ICON_FILE="$RESOURCES_DIR/$APP_NAME.icns"
  iconutil -c icns "$ICONS_DIR/icon.iconset" -o "$ICON_FILE"
elif [[ -f "$ICONS_DIR/icon.png" ]]; then
  ICON_FILE="$RESOURCES_DIR/$APP_NAME.icns"
  mkdir -p "$RESOURCES_DIR/tmp.iconset"
  sips -z 16 16 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_16x16.png" >/dev/null 2>&1 || true
  sips -z 32 32 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_16x16@2x.png" >/dev/null 2>&1 || true
  sips -z 32 32 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_32x32.png" >/dev/null 2>&1 || true
  sips -z 64 64 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_32x32@2x.png" >/dev/null 2>&1 || true
  sips -z 128 128 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_128x128.png" >/dev/null 2>&1 || true
  sips -z 256 256 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_128x128@2x.png" >/dev/null 2>&1 || true
  sips -z 256 256 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_256x256.png" >/dev/null 2>&1 || true
  sips -z 512 512 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_256x256@2x.png" >/dev/null 2>&1 || true
  sips -z 512 512 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_512x512.png" >/dev/null 2>&1 || true
  sips -z 1024 1024 "$ICONS_DIR/icon.png" --out "$RESOURCES_DIR/tmp.iconset/icon_512x512@2x.png" >/dev/null 2>&1 || true
  iconutil -c icns "$RESOURCES_DIR/tmp.iconset" -o "$ICON_FILE"
  rm -rf "$RESOURCES_DIR/tmp.iconset"
fi

echo "Building $APP_NAME for darwin/$GOARCH_VALUE..."
CGO_ENABLED=1 GOOS=darwin GOARCH="$GOARCH_VALUE" go build -trimpath -ldflags='-s -w' -o "$MACOS_DIR/$APP_NAME" ./cmd

cat > "$PLIST_PATH" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>$APP_NAME</string>
  <key>CFBundleIdentifier</key>
  <string>com.local.$APP_NAME</string>
  <key>CFBundleName</key>
  <string>$APP_NAME</string>
  <key>CFBundleDisplayName</key>
  <string>$APP_NAME</string>
  <key>CFBundleVersion</key>
  <string>1.0</string>
  <key>CFBundleShortVersionString</key>
  <string>1.0</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>LSMinimumSystemVersion</key>
  <string>10.15</string>
  <key>NSHighResolutionCapable</key>
  <true/>
  <key>CFBundleIconFile</key>
  <string>$(basename "$ICON_FILE" .icns)</string>
</dict>
</plist>
EOF

if [[ -n "$ICON_FILE" ]]; then
  cp "$ICON_FILE" "$RESOURCES_DIR/$(basename "$ICON_FILE")"
fi

chmod +x "$MACOS_DIR/$APP_NAME"

echo "Built app bundle: $BUNDLE_PATH"
