// +build windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath, carinaBinaryName string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to %s/\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("\"%v\"\n", path.Join(basepath, "docker.cmd"))
	s += fmt.Sprintf("# Run the command above to set your docker environment")
	return s
}
