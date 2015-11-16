// +build !windows

package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
)

func sourceHelpString(basepath string, name string) string {
	s := "#\n"
	s += fmt.Sprintf("# Credentials written to \"%s\"\n", basepath)
	s += "#\n"
	s += fmt.Sprintf("eval `carina env %v`\n", name)
	s += fmt.Sprintf("# Run the command above to get your Docker environment variables set\n")
	return s
}

const defaultDotDir = ".carina"
const defaultNonDotDir = "carina"
const xdgDataHomeEnvVar = "XDG_DATA_HOME"

func userHomeDir() (string, error) {
	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME"), nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.HomeDir, nil
}

// CarinaCredentialsBaseDir get the current base directory for carina credentials
func CarinaCredentialsBaseDir() (string, error) {
	if os.Getenv(CarinaHomeDirEnvVar) != "" {
		return os.Getenv(CarinaHomeDirEnvVar), nil
	}
	if os.Getenv(CredentialsBaseDirEnvVar) != "" {
		return os.Getenv(CredentialsBaseDirEnvVar), nil
	}

	// Support XDG
	if os.Getenv(xdgDataHomeEnvVar) != "" {
		return path.Join(os.Getenv(xdgDataHomeEnvVar), defaultNonDotDir), nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, defaultDotDir), nil
}
