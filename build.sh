#!/bin/bash
# Kora Engine — Build Script
# Uso: ./build.sh [debug|release]

set -e

MODE="${1:-release}"

echo "======================================"
echo "Kora Engine Build"
echo "======================================"
echo ""
echo "Modo: $MODE"
echo ""

# Verificar Go
if ! command -v go &> /dev/null; then
    echo "Erro: Go não encontrado. Instale Go 1.22+ e tente novamente."
    exit 1
fi

# Verificar ANDROID_HOME
if [ -z "$ANDROID_HOME" ]; then
    echo "Aviso: ANDROID_HOME não definido. Usando padrão..."
    if [ -d "$HOME/Android/Sdk" ]; then
        export ANDROID_HOME="$HOME/Android/Sdk"
    else
        echo "Erro: Android SDK não encontrado em $HOME/Android/Sdk"
        exit 1
    fi
fi

echo "ANDROID_HOME: $ANDROID_HOME"
echo "GO: $(go version)"
echo ""

# Ler valores do android/app/build.gradle
APP_NAME=$(grep -oP '(?<=applicationId ").*(?=")' android/app/build.gradle)
VERSION=$(grep -oP '(?<=versionName ")[^"]+' android/app/build.gradle)

BUILD_DIR="android/app/build/outputs/apk/$MODE"
APK_NAME="$APP_NAME-$MODE-v$VERSION.apk"

echo "======================================"
echo "Buildando APK..."
echo "======================================"
echo ""

cd android
./build.sh "$MODE"

echo ""
echo "======================================"
echo "Build Concluído!"
echo "======================================"
echo ""
echo "APK gerado: $BUILD_DIR/$APK_NAME"
echo ""
echo "Para instalar no dispositivo:"
echo "  adb install $BUILD_DIR/$APK_NAME"
echo ""
echo "Para depuração:"
echo "  adb shell monkey -p $APP_NAME -c android.intent.category.LAUNCHER 1"
echo ""
