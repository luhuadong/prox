#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$project_dir/internal/shell/completion.bash"

COMP_WORDS=(prox co)
COMP_CWORD=1
_prox_completion
[[ " ${COMPREPLY[*]} " == *' config '* ]]
[[ " ${COMPREPLY[*]} " == *' completion '* ]]

COMP_WORDS=(prox config val)
COMP_CWORD=2
_prox_completion
[[ "${COMPREPLY[*]}" == validate ]]

COMP_WORDS=(prox config init --p)
COMP_CWORD=3
_prox_completion
[[ "${COMPREPLY[*]}" == --proxy ]]

COMP_WORDS=(prox check --l)
COMP_CWORD=2
_prox_completion
[[ "${COMPREPLY[*]}" == --local ]]

printf 'completion tests passed\n'
