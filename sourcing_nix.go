// +build !windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath string, command string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to \"%s\"/\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("source \"%v\"\n", path.Join(basepath, "docker.env"))
	s += fmt.Sprintf("# Run the command above or eval a subshell like so\n")
	s += fmt.Sprintf("#   eval \"$( %v )\"", command)
	return s
}
