# Bash completion for prox.

_prox_completion() {
    local current previous command config_command index
    current=${COMP_WORDS[COMP_CWORD]}
    previous=''
    if (( COMP_CWORD > 0 )); then
        previous=${COMP_WORDS[COMP_CWORD - 1]}
    fi

    if [[ $previous == --config ]]; then
        mapfile -t COMPREPLY < <(compgen -f -- "$current")
        return
    fi

    command=''
    index=1
    while (( index < COMP_CWORD )); do
        case "${COMP_WORDS[index]}" in
            --config)
                ((index += 2))
                ;;
            --config=*)
                ((index += 1))
                ;;
            *)
                command=${COMP_WORDS[index]}
                break
                ;;
        esac
    done

    case "$command" in
        '')
            mapfile -t COMPREPLY < <(compgen -W '--config init config on off status check run completion version help' -- "$current")
            ;;
        init|completion)
            mapfile -t COMPREPLY < <(compgen -W 'bash' -- "$current")
            ;;
        config)
            config_command=''
            ((index += 1))
            if (( index < COMP_CWORD )); then
                config_command=${COMP_WORDS[index]}
            fi
            case "$config_command" in
                '')
                    mapfile -t COMPREPLY < <(compgen -W 'path show init validate help' -- "$current")
                    ;;
                init)
                    mapfile -t COMPREPLY < <(compgen -W '--proxy --force --help' -- "$current")
                    ;;
            esac
            ;;
        check)
            mapfile -t COMPREPLY < <(compgen -W '--local --url --quiet --help' -- "$current")
            ;;
        run)
            if [[ $previous == -- ]]; then
                mapfile -t COMPREPLY < <(compgen -c -- "$current")
            fi
            ;;
    esac
}

complete -F _prox_completion prox
