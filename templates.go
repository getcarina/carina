package main

// BashCompletionTemplate is a basic template we'll start off with
var BashCompletionTemplate = `
#!/usr/bin/env bash

# bash completion for kingpin programs

__kingpin_{{.App.Name}}()
{
  local cur prev opts
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  local subcommands flags globalflags

  # We only go one deep, plus flags
  subcommands='{{range .App.FlattenedCommands}}{{if not .Hidden}}\
{{.Name}} {{end}}{{end}}'

  globalflags='{{range .App.Flags}}{{if not .Hidden}}\
--{{.Name}} {{end}}{{end}}'

  if [ "$prev" == "{{.App.Name}}" ]; then
    opts="$subcommands $globalflags"
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
  fi

  case $prev in
{{range .App.FlattenedCommands}}\
  "{{.Name}}")
    flags='{{range .Flags}}{{if not .Hidden}} --{{.Name}}{{end}}{{end}}'
    opts="$flags $globalflags"
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
    ;;
{{end}}\
  esac
}

# complete is a bash builtin, but recent versions of ZSH come with a function
# called bashcompinit that will create a complete in ZSH. If the user is in
# ZSH, load and run bashcompinit before calling the complete function.
if [[ -n ${ZSH_VERSION-} ]]; then
	autoload -U +X bashcompinit && bashcompinit
fi

complete -o default -o nospace -F __kingpin_{{.App.Name}} {{.App.Name}}
`
