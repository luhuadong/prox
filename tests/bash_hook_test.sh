#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
temporary_dir="$(mktemp -d)"
trap 'rm -rf "$temporary_dir"' EXIT

cat >"$temporary_dir/prox" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

config_path=''
if [[ ${1:-} == --config ]]; then
    config_path="${2:-}"
    shift 2
elif [[ ${1:-} == --config=* ]]; then
    config_path="${1#--config=}"
    shift
fi

case "${1:-}" in
    __shell-env)
        if [[ -n $config_path ]]; then
            printf '%s\n' "$config_path" >"${PROX_STUB_CONFIG_LOG:?}"
        fi
        if [[ ${PROX_STUB_FAIL:-0} == 1 ]]; then
            printf 'stub endpoint unavailable\n' >&2
            exit 1
        fi
        cat <<'ENV'
export http_proxy='http://127.0.0.1:7890'
export https_proxy='http://127.0.0.1:7890'
export all_proxy='http://127.0.0.1:7890'
export HTTPS_PROXY='http://127.0.0.1:7890'
export ALL_PROXY='http://127.0.0.1:7890'
export no_proxy='localhost,127.0.0.1,::1,.local'
export NO_PROXY='localhost,127.0.0.1,::1,.local'
__PROX_PROXY_DISPLAY='http://127.0.0.1:7890'
ENV
        ;;
    status)
        printf 'State: %s\n' "${PROX_INTERNAL_STATE:-inactive}"
        ;;
    *)
        printf 'unexpected stub command: %s\n' "$*" >&2
        exit 2
        ;;
esac
STUB
chmod +x "$temporary_dir/prox"
PATH="$temporary_dir:$PATH"
export PROX_STUB_CONFIG_LOG="$temporary_dir/config.log"

unset http_proxy https_proxy all_proxy HTTPS_PROXY ALL_PROXY no_proxy NO_PROXY
export http_proxy='http://original.example:8080'
no_proxy='original.local'
original_no_proxy_declaration="$(declare -p no_proxy)"

# The embedded file contains only one version placeholder and is valid Bash as-is.
source "$project_dir/internal/shell/prox.bash"

# A failed pre-check must leave the current environment untouched.
export PROX_STUB_FAIL=1
if prox on >/dev/null 2>/dev/null; then
    printf 'prox on unexpectedly succeeded\n' >&2
    exit 1
fi
unset PROX_STUB_FAIL
[[ "$http_proxy" == 'http://original.example:8080' ]]
[[ ! -v https_proxy ]]

prox on >/dev/null
[[ "$http_proxy" == 'http://127.0.0.1:7890' ]]
[[ "$https_proxy" == 'http://127.0.0.1:7890' ]]

# Repeated activation is idempotent and must not replace the original snapshot.
prox on >/dev/null

https_proxy='http://drifted.example:8080'
status_output="$(prox status)"
[[ "$status_output" == *'State: drifted'* ]]

prox off >/dev/null 2>/dev/null
[[ "$http_proxy" == 'http://original.example:8080' ]]
[[ ! -v https_proxy ]]
[[ "$(declare -p no_proxy)" == "$original_no_proxy_declaration" ]]
[[ ! -v ALL_PROXY ]]

# Repeated deactivation is also idempotent.
prox off >/dev/null

# Global configuration options must still pass through the Shell wrapper.
prox --config "$temporary_dir/custom.json" on >/dev/null
[[ "$(<"$PROX_STUB_CONFIG_LOG")" == "$temporary_dir/custom.json" ]]
status_output="$(prox --config="$temporary_dir/custom.json" status)"
[[ "$status_output" == *'State: active'* ]]
prox off >/dev/null

printf 'bash hook tests passed\n'
