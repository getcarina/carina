// +build !windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath, carinaBinaryName string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to %s/\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("source \"%v\"\n", path.Join(basepath, "docker.env"))
	s += fmt.Sprintf("# Run the command above or eval a subshell with your arguments to %v\n", carinaBinaryName)
	s += fmt.Sprintf("#   eval \"$( %v command... )\"", carinaBinaryName)
	return s
}
