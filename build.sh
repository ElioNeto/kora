#!/bin/bash
# Kora Engine — Build Script
# Usage: ./build.sh [debug|release]
#
# Builds the Kora game engine and creates an Android APK/AAB.
# Requires Go 1.22+, Android SDK, and gomobile.

set -euo pipefail

MODE="${1:-release}"
APP_PKG="github.com/ElioNeto/kora/cmd/kora-android"

echo "======================================"
echo "Kora Engine Build"
echo "======================================"
echo ""
echo "Mode: $MODE"
echo ""

# Check Go
if ! command -v go &>/dev/null; then
    echo "ERROR: Go not found. Install Go 1.22+ and try again."
    exit 1
fi

# Check ANDROID_HOME
if [ -z "${ANDROID_HOME:-}" ]; then
    if [ -d "$HOME/Android/Sdk" ]; then
        export ANDROID_HOME="$HOME/Android/Sdk"
    elif [ -d "$HOME/android" ]; then
        export ANDROID_HOME="$HOME/android"
    else
        echo "ERROR: ANDROID_HOME not set and no SDK found in default locations."
        echo "  Install Android SDK or set ANDROID_HOME explicitly."
        exit 1
    fi
fi

# Check gomobile
if ! command -v gomobile &>/dev/null; then
    echo "gomobile not found. Installing..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
fi

# Check NDK
if [ -z "${ANDROID_NDK_HOME:-}" ]; then
    NDK_SEARCH=$(ls -d "$ANDROID_HOME/ndk/"*/ 2>/dev/null | head -1 || true)
    if [ -n "$NDK_SEARCH" ]; then
        export ANDROID_NDK_HOME="$NDK_SEARCH"
        echo "  Using NDK: $ANDROID_NDK_HOME"
    else
        echo "WARNING: ANDROID_NDK_HOME not set. gomobile may fail."
        echo "  Install NDK via SDK Manager or set ANDROID_NDK_HOME."
    fi
fi

echo "ANDROID_HOME: $ANDROID_HOME"
echo "GO: $(go version)"
echo ""

echo "======================================"
echo "Building APK..."
echo "======================================"
echo ""

cd "$(dirname "$0")"
./android/build.sh "$MODE"

echo ""
echo "======================================"
echo "Build Complete!"
echo "======================================"
echo ""
echo "Output: build/android/"
echo ""
echo "Install on device:"
echo "  adb install build/android/kora-debug.apk"
echo ""
