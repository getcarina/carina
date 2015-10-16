// +build !windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath, carinaBinaryName string) string {
	s := ""
	s += fmt.Sprintf("source \"%v\"\n", path.Join(basepath, "docker.env"))
	s += fmt.Sprintf("# Run the above or eval a subshell with your arguments to %v\n", carinaBinaryName)
	s += fmt.Sprintf("# eval \"$( %v command... )\" \n", carinaBinaryName)
	return s
}
