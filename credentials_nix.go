// +build !windows

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func credentialsNextStepsString(clusterName string) string {
	return fmt.Sprintf("# To see how to connect to your cluster, run: carina env %s\n", clusterName)
}

func getCredentialFilePath(basepath string, shell string) string {
	return filepath.Join(basepath, "docker.env")
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
		return "", err
	}
	return currentUser.HomeDir, nil
}
