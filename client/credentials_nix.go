// +build !windows

package client

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/user"
	"path/filepath"
)

func CredentialsNextStepsString(clusterName string) string {
	return fmt.Sprintf("# To see how to connect to your cluster, run: carina env %s\n", clusterName)
}

func getCredentialFilePath(basepath string, shell string) string {
	switch shell {
	case "fish":
		return filepath.Join(basepath, "docker.fish")
	default:
		return filepath.Join(basepath, "docker.env")
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
