#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/cmd/wails"
FRONTEND_DIR="$ROOT_DIR/cmd/wails/frontend"

TARGET_PLATFORM="${1:-windows/amd64}"

usage() {
  cat <<USAGE
Usage:
  $(basename "$0") [windows/amd64|windows/arm64]

Examples:
  $(basename "$0")
  $(basename "$0") windows/arm64
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "[ERROR] Missing required command: $1" >&2
    exit 1
  fi
}

resolve_wails() {
  if command -v wails >/dev/null 2>&1; then
    command -v wails
    return
  fi
  local go_bin
  go_bin="$(go env GOPATH 2>/dev/null)/bin/wails"
  if [[ -x "$go_bin" ]]; then
    echo "$go_bin"
    return
  fi
  echo ""
}

WAILS_BIN="$(resolve_wails)"
if [[ -z "$WAILS_BIN" ]]; then
  echo "[ERROR] Wails CLI not found. Install with:" >&2
  echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
  echo "[HINT] Repository root contains './wails' app binary, which is not the Wails CLI." >&2
  exit 1
fi

require_cmd pnpm

if [[ "$TARGET_PLATFORM" != "windows/amd64" && "$TARGET_PLATFORM" != "windows/arm64" ]]; then
  echo "[ERROR] Unsupported target platform: $TARGET_PLATFORM" >&2
  usage
  exit 1
fi

if [[ "$OSTYPE" == linux* ]]; then
  if [[ "$TARGET_PLATFORM" == "windows/amd64" ]] && ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then
    echo "[WARN] x86_64-w64-mingw32-gcc not found. Cross-compile may fail on Linux." >&2
  fi
  if [[ "$TARGET_PLATFORM" == "windows/arm64" ]] && ! command -v aarch64-w64-mingw32-gcc >/dev/null 2>&1; then
    echo "[WARN] aarch64-w64-mingw32-gcc not found. Cross-compile may fail on Linux." >&2
  fi
fi

if [[ -z "${GOFLAGS:-}" ]]; then
  export GOFLAGS="-buildvcs=false"
elif [[ " ${GOFLAGS} " != *" -buildvcs=false "* ]]; then
  export GOFLAGS="${GOFLAGS} -buildvcs=false"
fi

if [[ -z "${GOCACHE:-}" ]]; then
  export GOCACHE="/tmp/go-build-cache"
fi
mkdir -p "$GOCACHE"

echo "[1/3] Installing frontend dependencies..."
cd "$FRONTEND_DIR"
pnpm install --frozen-lockfile

echo "[2/3] Building frontend assets..."
pnpm build

BUILD_TAGS="wails,desktop,production"

echo "[3/3] Building Wails app for $TARGET_PLATFORM ..."
cd "$APP_DIR"
"$WAILS_BIN" build -tags "$BUILD_TAGS" -platform "$TARGET_PLATFORM" -clean

echo "[DONE] Build finished. Check output in: $APP_DIR/build/bin"
