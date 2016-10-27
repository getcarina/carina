// +build windows

package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CredentialsNextStepsString returns instructions to load the cluster credentials
func CredentialsNextStepsString(clusterName string) string {
	return fmt.Sprintf("# To see how to connect to your cluster, run: carina env %s --shell cmd|powershell|bash\n", clusterName)
}

func getCredentialScriptPath(basepath string, shell string) (string, error) {
	scriptPrefix, err := getCredentialScriptPrefix(basepath)
	if err != nil {
		return "", err
	}

	pathPrefix := filepath.Join(basepath, scriptPrefix)

	switch shell {
	case "powershell":
		return pathPrefix + ".ps1", nil
	case "cmd":
		return pathPrefix + ".cmd", nil
	case "bash":
		return pathPrefix + ".env", nil
	default:
		return "", fmt.Errorf("Invalid shell specified: %s", shell)
	}
}

func forceUnixPath(winPath string) string {
	// Convert C:/ --> /C/
	unixPath := "/" + strings.Replace(winPath, ":\\", "/", 1)
	// Replace path seperators
	unixPath = strings.Replace(unixPath, "\\", "/", -1)
	return unixPath
}

func sourceHelpString(credentialFile string, clusterName string, shell string) string {
	switch shell {
	case "powershell":
		s := fmt.Sprintf(". %s\n", credentialFile)
		s += fmt.Sprintf("# Run the command below to load environment variables for docker or kubectl:\n")
		s += fmt.Sprintf("# carina env %s --shell powershell | iex", clusterName) // PowerShell bombs if you have an empty line, leaving out
		return s
	case "cmd":
		s := fmt.Sprintf("# Run the command below to load environment variables for docker or kubectl:\n")
		s += fmt.Sprintf("CALL %s\n", credentialFile)
		return s
	default: // Windows Bash
		s := fmt.Sprintf("source %s\n", forceUnixPath(credentialFile))
		s += fmt.Sprintf("# Run the command below to load environment variables for docker or kubectl:\n")
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
