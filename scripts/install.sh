#!/usr/bin/env sh

set -eu

OWNER_REPO="rohankmr414/grove"
BINARY_NAME="grove"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${1:-${VERSION:-latest}}"
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need_cmd uname
need_cmd mktemp
need_cmd tar

if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
else
  echo "missing required command: curl or wget" >&2
  exit 1
fi

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Linux) os="Linux" ;;
  Darwin) os="Darwin" ;;
  *)
    echo "unsupported operating system: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

fetch() {
  # shellcheck disable=SC2086
  $FETCH "$1"
}

installed_version() {
  if [ ! -x "$BINARY_PATH" ]; then
    return 1
  fi

  version_output="$("$BINARY_PATH" version 2>/dev/null || true)"
  version_line="$(printf '%s\n' "$version_output" | sed -n '1s/^grove[[:space:]]\+//p')"
  if [ -z "$version_line" ]; then
    return 1
  fi

  printf '%s\n' "$version_line"
}

resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    printf '%s\n' "${VERSION#v}"
    return
  fi

  latest_json="$(fetch "https://api.github.com/repos/$OWNER_REPO/releases/latest")"
  tag="$(printf '%s' "$latest_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
  if [ -z "$tag" ]; then
    echo "failed to resolve latest release tag" >&2
    exit 1
  fi
  printf '%s\n' "${tag#v}"
}

version="$(resolve_version)"
archive="${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$OWNER_REPO/releases/download/v${version}/${archive}"

current_version="$(installed_version || true)"
if [ -n "$current_version" ]; then
  if [ "$current_version" = "$version" ]; then
    echo "$BINARY_NAME $version is already installed at $BINARY_PATH"
    exit 0
  fi
  echo "Upgrading $BINARY_NAME from $current_version to $version"
else
  echo "Installing $BINARY_NAME $version to $INSTALL_DIR"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

echo "Downloading $url"
fetch "$url" >"$tmpdir/$archive"

mkdir -p "$INSTALL_DIR"
tar -xzf "$tmpdir/$archive" -C "$tmpdir"

if command -v install >/dev/null 2>&1; then
  install -m 0755 "$tmpdir/$BINARY_NAME" "$BINARY_PATH"
else
  cp "$tmpdir/$BINARY_NAME" "$BINARY_PATH"
  chmod 0755 "$BINARY_PATH"
fi

shell_name="${SHELL##*/}"

echo "Installed $BINARY_PATH"
echo
echo "Shell integration:"
case "$shell_name" in
  zsh|bash)
    echo "  Current shell:"
    echo "    eval \"\$($BINARY_PATH shell-init $shell_name)\""
    echo
    echo "  Permanent setup:"
    echo "    echo 'eval \"\$($BINARY_PATH shell-init $shell_name)\"' >> \$HOME/.${shell_name}rc"
    ;;
  *)
    echo "  Run one of these in your shell:"
    echo "    eval \"\$($BINARY_PATH shell-init zsh)\""
    echo "    eval \"\$($BINARY_PATH shell-init bash)\""
    ;;
esac
