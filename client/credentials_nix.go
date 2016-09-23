// +build !windows

package client

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
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
	s += fmt.Sprintf("# Run the command below to get your Docker environment variables set:\n")
	s += fmt.Sprintf("# eval $(carina env %s)", clusterName)
	return s
}

func userHomeDir() (string, error) {
	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME"), nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve the current user")
	}
	return currentUser.HomeDir, nil
}
