#!/usr/bin/env bash
# =============================================================================
# Kora Android — one-time development environment setup
# =============================================================================
# Run this once after cloning the repo to install Go mobile tooling.
# Requires: Go 1.22+, Android SDK with NDK.
# =============================================================================

set -euo pipefail

echo "▶ Installing gomobile..."
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest

echo "▶ Running gomobile init (downloads Android NDK toolchain)..."
gomobile init

echo "▶ Tidying Go modules..."
go mod tidy

echo ""
echo "✅ Setup complete. Run builds with:"
echo "   ./android/build.sh debug    # APK for sideloading"
echo "   ./android/build.sh release  # Signed APK for Google Play"
