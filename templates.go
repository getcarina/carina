package main

// BashCompletionTemplate is a basic template we'll start off with
var BashCompletionTemplate = `
#!/usr/bin/env bash

# bash completion for kingpin programs

__kingpin_generate_completion()
{
  declare current_word
  current_word="${COMP_WORDS[COMP_CWORD]}"
  COMPREPLY=($(compgen -W "$1" -- "$current_word"))
  return 0
}

__kingpin_commands ()
{
  declare current_word
  declare command

  current_word="${COMP_WORDS[COMP_CWORD]}"

  COMMANDS='\
    {{range .App.FlattenedCommands}}\
    {{if not .Hidden}}\
{{.Name}}\
    {{end}}\
    {{end}}\
    '

    if [ ${#COMP_WORDS[@]} == 4 ]; then

      command="${COMP_WORDS[COMP_CWORD-2]}"
      case "${command}" in
      # If commands have subcommands
      esac

    else

      case "${current_word}" in
      -*)     __kingpin_options ;;
      *)      __kingpin_generate_completion "$COMMANDS" ;;
      esac

    fi
}

__kingpin_options ()
{
  OPTIONS=''
  __kingpin_generate_completion "$OPTIONS"
}

__kingpin ()
{
  declare previous_word
  previous_word="${COMP_WORDS[COMP_CWORD-1]}"

  case "$previous_word" in
  *)              __kingpin_commands ;;
  esac

  return 0
}

# complete is a bash builtin, but recent versions of ZSH come with a function
# called bashcompinit that will create a complete in ZSH. If the user is in
# ZSH, load and run bashcompinit before calling the complete function.
if [[ -n ${ZSH_VERSION-} ]]; then
	autoload -U +X bashcompinit && bashcompinit
fi

complete -o default -o nospace -F __kingpin {{.App.Name}}
`
