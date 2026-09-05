#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
temporary_dir="$(mktemp -d)"
trap 'rm -rf "$temporary_dir"' EXIT

make -C "$project_dir" --no-print-directory install \
    DESTDIR="$temporary_dir/root" PREFIX=/usr >/dev/null
[[ -x "$temporary_dir/root/usr/bin/prox" ]]

case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) printf 'unsupported test architecture\n' >&2; exit 1 ;;
esac

version=9.8.7
archive="prox_${version}_linux_${arch}.tar.gz"
archive_root="$temporary_dir/archive/prox_${version}_linux_${arch}"
release_dir="$temporary_dir/releases/v${version}"
mkdir -p "$archive_root" "$release_dir"
install -m 0755 "$project_dir/dist/prox" "$archive_root/prox"
tar -C "$temporary_dir/archive" -czf "$release_dir/$archive" "prox_${version}_linux_${arch}"
(cd "$release_dir" && sha256sum "$archive" >checksums.txt)

PROX_INSTALL_BASE_URL="file://$temporary_dir/releases" \
    sh "$project_dir/scripts/install.sh" \
    --version "v$version" \
    --bin-dir "$temporary_dir/user-bin" >/dev/null

[[ -x "$temporary_dir/user-bin/prox" ]]
cmp "$project_dir/dist/prox" "$temporary_dir/user-bin/prox"

printf 'install tests passed\n'
