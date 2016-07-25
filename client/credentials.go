package client

import (
	"os"
	"path/filepath"
)

const clusterDirName = "clusters"
const defaultDotDir = ".carina"
const defaultNonDotDir = "carina"
const xdgDataHomeEnvVar = "XDG_DATA_HOME"

// CredentialsBaseDirEnvVar environment variable name for where credentials are downloaded to by default
const CredentialsBaseDirEnvVar = "CARINA_CREDENTIALS_DIR"

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


func buildClusterCredentialsPath(userName string, clusterName string, customPath string) (string, error) {
	var credentialsPath string

	// Use the default path, if the user didn't specify a special path where the credentials are stored
	if customPath == "" {
		var baseDir string
		baseDir, err := CarinaCredentialsBaseDir()
		if err != nil {
			return "", err
		}
		credentialsPath = filepath.Join(baseDir, clusterDirName, userName, clusterName)
	}

	credentialsPath = filepath.Clean(credentialsPath)
	return credentialsPath, nil
}
