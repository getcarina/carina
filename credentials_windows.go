// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func sourceHelpString(basepath string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to %s/\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("\"%v\"\n", path.Join(basepath, "docker.cmd"))
	s += fmt.Sprintf("# Run the command above to set your docker environment")
	return s
}

func userHomeDir() (string, error) {
	if os.Getenv("HOMEDRIVE") != "" && os.Getenv("HOMEPATH") != "" {
		return path.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH")), nil
	}
	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME"), nil
	}

	return "", errors.New("Unable to locate home directory")
}

// CarinaCredentialsBaseDir get the current base directory for carina credentials
func CarinaCredentialsBaseDir() (string, error) {
	if os.Getenv(CredentialsBaseDirEnvVar) != "" {
		return os.Getenv(CredentialsBaseDirEnvVar), nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, "carina"), nil
}
