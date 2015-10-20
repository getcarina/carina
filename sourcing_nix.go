// +build !windows

package main

import (
	"fmt"
	"path"
)

func sourceHelpString(basepath string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to \"%s\"/\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("source \"%v\"\n", path.Join(basepath, "docker.env"))
	s += fmt.Sprintf("# Run the command above to get your Docker environment variables set\n")
	return s
}
