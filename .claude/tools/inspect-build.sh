#!/usr/bin/env bash
# inspect-build.sh
# Ferramenta de leitura do estado do build local do Kora Engine.
# Uso: bash .claude/tools/inspect-build.sh
# Retorna JSON com estado do compilador, binários e artefatos Android.
# Seguro: apenas leitura, sem efeitos colaterais.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# ---- helpers ----
file_info() {
  local f="$1"
  if [ -f "$f" ]; then
    local size mod
    size=$(du -sh "$f" 2>/dev/null | cut -f1)
    mod=$(date -r "$f" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || stat -c '%y' "$f" 2>/dev/null | cut -d'.' -f1)
    echo "{\"exists\": true, \"path\": \"$f\", \"size\": \"$size\", \"modified\": \"$mod\"}"
  else
    echo "{\"exists\": false, \"path\": \"$f\"}"
  fi
}

dir_info() {
  local d="$1"
  if [ -d "$d" ]; then
    local size count
    size=$(du -sh "$d" 2>/dev/null | cut -f1)
    count=$(find "$d" -type f 2>/dev/null | wc -l | tr -d ' ')
    echo "{\"exists\": true, \"path\": \"$d\", \"size\": \"$size\", \"file_count\": $count}"
  else
    echo "{\"exists\": false, \"path\": \"$d\"}"
  fi
}

# ---- go version ----
GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' || echo "not found")
GO_MOD_MODULE=$(grep '^module' "$ROOT/go.mod" 2>/dev/null | awk '{print $2}' || echo "unknown")
GO_MOD_GO=$(grep '^go ' "$ROOT/go.mod" 2>/dev/null | awk '{print $2}' || echo "unknown")

# ---- binaries ----
BIN_COMPILER=$(file_info "$ROOT/bin/kora-compiler")
BIN_KORA=$(file_info "$ROOT/bin/kora")

# ---- generated ks.go files ----
KSGO_COUNT=$(find "$ROOT" -name '*.ks.go' -not -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
KSGO_LIST=$(find "$ROOT" -name '*.ks.go' -not -path '*/vendor/*' 2>/dev/null | head -20 | sed 's|'"$ROOT"'/||g' | awk '{printf "\"%s\",", $0}' | sed 's/,$//')

# ---- android artifacts ----
APK_COUNT=$(find "$ROOT" -name '*.apk' 2>/dev/null | wc -l | tr -d ' ')
AAB_COUNT=$(find "$ROOT" -name '*.aab' 2>/dev/null | wc -l | tr -d ' ')
ANDROID_BUILD=$(dir_info "$ROOT/android/app/build")
ANDROID_GRADLE=$(dir_info "$ROOT/android/.gradle")

# ---- dist ----
DIST=$(dir_info "$ROOT/dist")

# ---- ks files (scripts do jogo) ----
KS_COUNT=$(find "$ROOT/examples" "$ROOT/templates" -name '*.ks' 2>/dev/null | wc -l | tr -d ' ')

# ---- go test cache ----
TEST_CACHE=$(go env GOCACHE 2>/dev/null || echo "unknown")

# ---- last build attempt (from .git log) ----
LAST_COMMIT=$(git -C "$ROOT" log -1 --format='%h %s (%cr)' 2>/dev/null || echo "git not available")
DIRTY=$(git -C "$ROOT" diff --quiet 2>/dev/null && echo false || echo true)

# ---- output JSON ----
cat <<EOF
{
  "tool": "inspect-build",
  "root": "$ROOT",
  "timestamp": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "go": {
    "version": "$GO_VERSION",
    "module": "$GO_MOD_MODULE",
    "required": "$GO_MOD_GO",
    "test_cache": "$TEST_CACHE"
  },
  "binaries": {
    "kora_compiler": $BIN_COMPILER,
    "kora": $BIN_KORA
  },
  "generated_files": {
    "ks_go_count": $KSGO_COUNT,
    "ks_go_files": [$KSGO_LIST]
  },
  "kscript_sources": {
    "ks_count": $KS_COUNT
  },
  "android": {
    "apk_count": $APK_COUNT,
    "aab_count": $AAB_COUNT,
    "build_dir": $ANDROID_BUILD,
    "gradle_cache": $ANDROID_GRADLE
  },
  "dist": $DIST,
  "git": {
    "last_commit": "$LAST_COMMIT",
    "has_uncommitted_changes": $DIRTY
  }
}
EOF
