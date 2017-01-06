package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const clusterDirName = "clusters"
const defaultDotDir = ".carina"
const defaultNonDotDir = "carina"
const xdgDataHomeEnvVar = "XDG_DATA_HOME"

// GetCredentialsDir gets the carina home directory, e.g. ~/.carina
func GetCredentialsDir() (string, error) {
	if os.Getenv(CarinaHomeDirEnvVar) != "" {
		return os.Getenv(CarinaHomeDirEnvVar), nil
	}

	// Support XDG
	if os.Getenv(xdgDataHomeEnvVar) != "" {
		return filepath.Join(os.Getenv(xdgDataHomeEnvVar), defaultNonDotDir), nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return "", errors.New("Unable to default CARINA_HOME to ~/.carina. Set the CARINA_HOME environment variable")
	}
	return filepath.Join(homeDir, defaultDotDir), nil
}

func buildClusterCredentialsPath(account Account, clusterName string, customPath string) (string, error) {
	var credentialsPath string

	// Use the default path, if the user didn't specify a special path where the credentials are stored
	if customPath == "" {
		var baseDir string
		baseDir, err := GetCredentialsDir()
		if err != nil {
			return "", err
		}

		clusterPrefix, err := account.GetClusterPrefix()
		if err != nil {
			return "", err
		}
		credentialsPath = filepath.Join(baseDir, clusterDirName, clusterPrefix, clusterName)
	}

	credentialsPath = filepath.Clean(credentialsPath)
	return credentialsPath, nil
}

// getCredentialScriptPrefix looks at a credentials bundle and identifies the
// script prefix (e.g. docker or kubectl) used by the shell scripts
func getCredentialScriptPrefix(credsPath string) (string, error) {
	scriptPattern := filepath.Join(credsPath, "*.env")
	results, _ := filepath.Glob(scriptPattern)
	if len(results) == 0 {
		return "", fmt.Errorf("Invalid credentials bundle, could not find the bash script (*.env) in %s", credsPath)
	}
	if len(results) > 1 {
		return "", fmt.Errorf("Invalid credentials bundle, multiple bash scripts (*.env) found in %s", credsPath)
	}

	bashScriptName := filepath.Base(results[0])

	return strings.TrimSuffix(bashScriptName, ".env"), nil
}
