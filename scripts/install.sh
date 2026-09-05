#!/bin/sh
set -eu

repository="luhuadong/prox"
base_url="${PROX_INSTALL_BASE_URL:-https://github.com/$repository/releases/download}"
version="latest"
bin_dir="${HOME:?HOME must be set}/.local/bin"

usage() {
    cat <<'USAGE'
Install prox from a GitHub release.

Usage:
  install.sh [--version VERSION] [--bin-dir DIRECTORY]

Options:
  --version VERSION    Release to install, for example v0.2.0 (default: latest)
  --bin-dir DIRECTORY  Installation directory (default: ~/.local/bin)
  -h, --help           Show this help
USAGE
}

fail() {
    printf 'prox installer: %s\n' "$*" >&2
    exit 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

while [ "$#" -gt 0 ]; do
    case "$1" in
        --version)
            [ "$#" -ge 2 ] && [ -n "$2" ] || fail "--version requires a value"
            version="$2"
            shift 2
            ;;
        --version=*)
            version=${1#--version=}
            [ -n "$version" ] || fail "--version requires a value"
            shift
            ;;
        --bin-dir)
            [ "$#" -ge 2 ] && [ -n "$2" ] || fail "--bin-dir requires a value"
            bin_dir="$2"
            shift 2
            ;;
        --bin-dir=*)
            bin_dir=${1#--bin-dir=}
            [ -n "$bin_dir" ] || fail "--bin-dir requires a value"
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            fail "unknown option: $1"
            ;;
    esac
done

[ "$(uname -s)" = Linux ] || fail "only Linux is currently supported"

case "$(uname -m)" in
    x86_64|amd64)
        arch=amd64
        ;;
    aarch64|arm64)
        arch=arm64
        ;;
    *)
        fail "unsupported architecture: $(uname -m)"
        ;;
esac

require_command curl
require_command awk
require_command install
require_command mktemp
require_command sha256sum
require_command tar

if [ "$version" = latest ]; then
    latest_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$repository/releases/latest") || \
        fail "could not resolve the latest release"
    version=${latest_url##*/}
    [ -n "$version" ] && [ "$version" != latest ] || fail "could not resolve the latest release"
fi

tag=$version
case "$tag" in
    v*) ;;
    *) tag="v$tag" ;;
esac
release_version=${tag#v}
case "$release_version" in
    ''|*[!0-9A-Za-z.+-]*) fail "invalid release version: $version" ;;
esac
archive="prox_${release_version}_linux_${arch}.tar.gz"
release_url="$base_url/$tag"

temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/prox-install.XXXXXX")
trap 'rm -rf "$temporary_dir"' EXIT HUP INT TERM

printf 'Downloading prox %s for linux/%s...\n' "$tag" "$arch"
curl -fsSL "$release_url/$archive" -o "$temporary_dir/$archive" || fail "could not download $archive"
curl -fsSL "$release_url/checksums.txt" -o "$temporary_dir/checksums.txt" || fail "could not download checksums.txt"

expected_checksum=$(awk -v file="$archive" '$2 == file { print; found = 1 } END { if (!found) exit 1 }' "$temporary_dir/checksums.txt") || \
    fail "checksums.txt does not contain $archive"
printf '%s\n' "$expected_checksum" | (cd "$temporary_dir" && sha256sum -c -) >/dev/null || \
    fail "checksum verification failed for $archive"

tar -xzf "$temporary_dir/$archive" -C "$temporary_dir"
binary="$temporary_dir/prox_${release_version}_linux_${arch}/prox"
[ -f "$binary" ] || fail "release archive does not contain prox"

install -d -- "$bin_dir"
install -m 0755 -- "$binary" "$bin_dir/prox"

printf 'Installed prox %s to %s/prox\n' "$tag" "$bin_dir"
case ":${PATH:-}:" in
    *":$bin_dir:"*) ;;
    *) printf 'Add %s to PATH before using prox.\n' "$bin_dir" ;;
esac
printf 'Next: run `prox check --local`.\n'
printf 'For on/off support, add `eval "$(prox init bash)"` to ~/.bashrc.\n'
