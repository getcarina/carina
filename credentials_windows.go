// +build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func getCredentialFilePath(basepath string, shell string) string {
	switch shell {
	case "powershell":
		return filepath.Join(basepath, "docker.ps1")
	case "cmd":
		return filepath.Join(basepath, "docker.cmd")
	default:
		return filepath.Join(basepath, "docker.env")
	}
}

func sourceHelpString(credentialFile string, clusterName string, shell string) string {
	switch shell {
	case "powershell":
		s := fmt.Sprintf(". %s\n", credentialFile)
		s += fmt.Sprintf("# Run the command below to get your Docker environment variables set:\n")
		s += fmt.Sprintf("# carina env %s --shell powershell | iex", clusterName) // PowerShell bombs if you have an empty line, leaving out
		return s
	case "cmd":
		s := fmt.Sprintf("CALL %s\n", credentialFile)
		s += fmt.Sprintf("# Run the command below to get your Docker environment variables set:\n")
		s += fmt.Sprint("# copy and paste the above command into your command prompt\n")
		return s
	default:
		s := fmt.Sprintf("source %s\n", credentialFile)
		s += fmt.Sprintf("# Run the command below to get your Docker environment variables set:\n")
		s += fmt.Sprintf("# eval $(carina env %s)\n", clusterName)
		return s
	}
}

func userHomeDir() (string, error) {
	if os.Getenv("HOMEDRIVE") != "" && os.Getenv("HOMEPATH") != "" {
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH")), nil
	}
	if os.Getenv("HOME") != "" {
		return os.Getenv("HOME"), nil
	}

	return "", errors.New("Unable to locate home directory")
}

// CarinaCredentialsBaseDir get the current base directory for carina credentials
func CarinaCredentialsBaseDir() (string, error) {
	if os.Getenv(CarinaHomeDirEnvVar) != "" {
		return os.Getenv(CarinaHomeDirEnvVar), nil
	}
	if os.Getenv(CredentialsBaseDirEnvVar) != "" {
		return os.Getenv(CredentialsBaseDirEnvVar), nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "carina"), nil
}
