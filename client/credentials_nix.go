// +build !windows

package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CredentialsNextStepsString returns instructions to load the cluster credentials
func CredentialsNextStepsString(clusterName string) string {
	return fmt.Sprintf("# To see how to connect to your cluster, run: carina env %s\n", clusterName)
}

func getCredentialScriptPath(basepath string, shell string) (string, error) {
	scriptPrefix, err := getCredentialScriptPrefix(basepath)
	if err != nil {
		return "", err
	}

	pathPrefix := filepath.Join(basepath, scriptPrefix)

	switch shell {
	case "fish":
		return pathPrefix + ".fish", nil
	case "bash":
		return pathPrefix + ".env", nil
	default:
		return "", fmt.Errorf("Invalid shell specified: %s", shell)
	}
}

func sourceHelpString(credentialFile string, clusterName string, shell string) string {
	s := fmt.Sprintf("source %s\n", credentialFile)
	s += fmt.Sprintf("# Run the command below to load environment variables for docker or kubectl:\n")
	s += fmt.Sprintf("# eval $(carina env %s)", clusterName)
	return s
}

func userHomeDir() (string, error) {
	home := os.Getenv("HOME")
	if home != "" {
		return home, nil
	}

	return "", errors.New("Unable to locate home directory")
}
