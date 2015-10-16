// +build windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath, carinaBinaryName string) string {
	s := ""
	s += fmt.Sprintf("\"%v\"\n", path.Join(basepath, "docker.cmd"))
	s += fmt.Sprintf("# Run the above to set your docker environment\n")
	return s
}
