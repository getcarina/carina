package main

import (
	"os"
	"path/filepath"
)

const defaultDotDir = ".carina"
const defaultNonDotDir = "carina"
const xdgDataHomeEnvVar = "XDG_DATA_HOME"

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
		return filepath.Join(os.Getenv(xdgDataHomeEnvVar), defaultNonDotDir), nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, defaultDotDir), nil
}
