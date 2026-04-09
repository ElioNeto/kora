#!/usr/bin/env bash
# =============================================================================
# Kora Android Build Script
# =============================================================================
# Prerequisites:
#   - Go 1.22+
#   - Android SDK + NDK (set ANDROID_HOME and ANDROID_NDK_HOME)
#   - gomobile installed: go install golang.org/x/mobile/cmd/gomobile@latest
#   - gomobile init (run once after install)
#
# Usage:
#   ./android/build.sh [debug|release]
#
# Outputs:
#   build/android/kora-debug.apk    (debug, sideloadable)
#   build/android/kora-release.aab  (release, Google Play)
# =============================================================================

set -euo pipefail

BUILD_TYPE="${1:-debug}"
MODULE="github.com/ElioNeto/kora"
APP_PKG="${MODULE}/cmd/kora-android"
OUT_DIR="build/android"
APP_ID="com.example.koragame"
APP_NAME="Kora Game"
VERSION_CODE="1"
VERSION_NAME="1.0.0"
MIN_SDK="21"
TARGET_SDK="34"

mkdir -p "${OUT_DIR}"

echo "▶ Kora Android build [${BUILD_TYPE}]"
echo "  App:     ${APP_ID}"
echo "  Version: ${VERSION_NAME} (${VERSION_CODE})"
echo ""

# ---------------------------------------------------------------------------
# Verify tools
# ---------------------------------------------------------------------------
command -v gomobile >/dev/null 2>&1 || {
    echo "❌ gomobile not found. Run: go install golang.org/x/mobile/cmd/gomobile@latest"
    exit 1
}
command -v keytool >/dev/null 2>&1 || {
    echo "❌ keytool not found. Install a JDK."
    exit 1
}

# ---------------------------------------------------------------------------
# Debug build → APK
# ---------------------------------------------------------------------------
if [ "${BUILD_TYPE}" = "debug" ]; then
    echo "▶ Building debug APK..."
    gomobile build \
        -target android \
        -androidapi ${MIN_SDK} \
        -o "${OUT_DIR}/kora-debug.apk" \
        "${APP_PKG}"
    echo "✅ Debug APK: ${OUT_DIR}/kora-debug.apk"
    echo ""
    echo "Install on device:"
    echo "  adb install -r ${OUT_DIR}/kora-debug.apk"
fi

# ---------------------------------------------------------------------------
# Release build → AAB (Google Play)
# ---------------------------------------------------------------------------
if [ "${BUILD_TYPE}" = "release" ]; then
    KEYSTORE="${KORA_KEYSTORE:-android/release.keystore}"
    KEY_ALIAS="${KORA_KEY_ALIAS:-kora}"
    KEY_PASS="${KORA_KEY_PASS:?Set KORA_KEY_PASS env var}"
    STORE_PASS="${KORA_STORE_PASS:?Set KORA_STORE_PASS env var}"

    # Generate keystore if it doesn't exist.
    if [ ! -f "${KEYSTORE}" ]; then
        echo "▶ Generating release keystore at ${KEYSTORE}..."
        keytool -genkeypair \
            -keystore "${KEYSTORE}" \
            -alias "${KEY_ALIAS}" \
            -keyalg RSA \
            -keysize 2048 \
            -validity 10000 \
            -storepass "${STORE_PASS}" \
            -keypass "${KEY_PASS}" \
            -dname "CN=Kora Game, O=KoraEngine, C=BR"
        echo "✅ Keystore created."
    fi

    echo "▶ Building release AAB..."
    gomobile build \
        -target android \
        -androidapi ${MIN_SDK} \
        -o "${OUT_DIR}/kora-release.apk" \
        "${APP_PKG}"

    # Sign with apksigner (part of Android build-tools).
    BUILD_TOOLS_DIR="$(ls -d ${ANDROID_HOME}/build-tools/* | sort -V | tail -1)"
    "${BUILD_TOOLS_DIR}/apksigner" sign \
        --ks "${KEYSTORE}" \
        --ks-key-alias "${KEY_ALIAS}" \
        --ks-pass "pass:${STORE_PASS}" \
        --key-pass "pass:${KEY_PASS}" \
        --out "${OUT_DIR}/kora-release-signed.apk" \
        "${OUT_DIR}/kora-release.apk"

    rm "${OUT_DIR}/kora-release.apk"
    echo "✅ Signed APK: ${OUT_DIR}/kora-release-signed.apk"
    echo ""
    echo "Upload to Google Play:"
    echo "  Use the .apk above or convert to AAB with bundletool."
fi
