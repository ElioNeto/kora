# Kora — Android Export

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl |
| Android SDK | API 34 | Android Studio / sdkmanager |
| Android NDK | r26+ | sdkmanager `ndk;26.3.11579264` |
| JDK | 17+ | For `keytool` / `apksigner` |
| gomobile | latest | `go install golang.org/x/mobile/cmd/gomobile@latest` |

## First-time setup

```bash
# 1. Set environment variables
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.3.11579264
export PATH=$PATH:$ANDROID_HOME/platform-tools

# 2. Run the setup script (installs gomobile, runs gomobile init)
./android/setup.sh
```

## Build

```bash
# Debug APK — install directly on a device/emulator
./android/build.sh debug
adb install -r build/android/kora-debug.apk

# Release APK — signed, ready for Google Play
export KORA_KEY_PASS=yourpassword
export KORA_STORE_PASS=yourpassword
./android/build.sh release
```

## How it works

```
KScript (.ks)
    ↓  kora compiler
Go source (gen/*.go)
    ↓  go build
Native library (libkoragame.so)   ← one per ABI: arm64-v8a, armeabi-v7a, x86_64
    ↓  gomobile
APK / AAB
    └── lib/arm64-v8a/libkoragame.so
    └── lib/armeabi-v7a/libkoragame.so
    └── assets/   (sprites, audio, tilemaps)
    └── AndroidManifest.xml
```

## Targeting ABIs

gomobile automatically builds for all supported Android ABIs:
- `arm64-v8a`  — modern phones (primary target)
- `armeabi-v7a` — older devices
- `x86_64`     — emulators

To target a single ABI for faster debug builds:
```bash
gomobile build -target android/arm64 -o build/android/kora-arm64.apk ./cmd/kora-android
```

## Signing for Google Play

Google Play requires AAB (Android App Bundle). Convert the signed APK:
```bash
bundletool build-apks \
  --bundle=build/android/kora-release-signed.apk \
  --output=build/android/kora.apks \
  --ks=android/release.keystore \
  --ks-key-alias=kora \
  --ks-pass=pass:$KORA_STORE_PASS
```

## Minimum SDK

`minSdkVersion 21` (Android 5.0 Lollipop) — covers ~99% of active devices.

## Assets

Place game assets in `cmd/kora-android/assets/`. They are bundled into the APK
and accessible at runtime via `os.Open("assets/sprite.png")`
(Ebitengine handles the Android asset path automatically).
