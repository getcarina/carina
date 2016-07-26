package client

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/magnum"
	"github.com/getcarina/carina/makeswarm"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Client struct {
	cacheEnabled bool
	Cache        *Cache
	Error        error
}

// CarinaHomeDirEnvVar is the environment variable name for carina data, config, etc.
const CarinaHomeDirEnvVar = "CARINA_HOME"

const CloudMakeSwarm = "public"
const CloudMakeCOE = "make-coe"
const CloudMagnum = "private"

func NewClient(cacheEnabled bool) *Client {
	client := &Client{cacheEnabled: cacheEnabled}
	err := client.initCache()
	if err != nil {
		client.cacheEnabled = false
		client.Error = CacheUnavailableError{cause: err}
	}
	return client
}

func (client *Client) initCache() error {
	if client.cacheEnabled {
		bd, err := GetCredentialsDir()
		if err != nil {
			return err
		}
		err = os.MkdirAll(bd, 0777)
		if err != nil {
			return err
		}

		cacheName, err := defaultCacheFilename()
		if err != nil {
			return err
		}
		client.Cache, err = LoadCache(cacheName)
		return err
	}
	return nil
}

func (client *Client) buildContainerService(account *Account) (common.ClusterService, error) {
	switch account.CloudType {
	case CloudMakeSwarm:
		credentials, _ := account.Credentials.(makeswarm.UserCredentials)
		return &makeswarm.MakeSwarm{Credentials: credentials}, nil
	case CloudMagnum:
		credentials, _ := account.Credentials.(magnum.MagnumCredentials)
		return &magnum.Magnum{Credentials: credentials}, nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", account.CloudType)
	}
}

// GetQuotas retrieves the quotas set for the account
func (client *Client) GetQuotas(account *Account) (common.Quotas, error) {
	var quotas common.Quotas

	svc, err := client.buildContainerService(account)
	if err != nil {
		return quotas, err
	}

	return svc.GetQuotas()
}

// CreateCluster creates a new cluster and prints the cluster information
func (client *Client) CreateCluster(account *Account, name string, nodes int, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.CreateCluster(name, nodes)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// GetClusterCredentials downloads the TLS certificates and configuration scripts for a cluster
func (client *Client) DownloadClusterCredentials(account *Account, name string, customPath string) (credentialsPath string, err error) {
	svc, err := client.buildContainerService(account)

	creds, err := svc.GetClusterCredentials(name)
	if err != nil {
		return "", err
	}

	username := account.Credentials.GetUserName()
	credentialsPath, err = buildClusterCredentialsPath(username, name, customPath)
	if err != nil {
		return "", errors.Wrap(err, "Unable to save downloaded cluster credentials")
	}

	// Ensure the credentials destination directory exists
	if credentialsPath != "." {
		err = os.MkdirAll(credentialsPath, 0777)
		if err != nil {
			return "", err
		}
	}

	for file, fileContents := range creds.Files {
		file = filepath.Join(credentialsPath, file)
		err = ioutil.WriteFile(file, fileContents, 0600)
		if err != nil {
			return "", err
		}
	}

	return credentialsPath, nil
}

// GetSourceCommand returns the shell command and appropriate help text to load a cluster's credentials
func (client *Client) GetSourceCommand(account *Account, shell string, name string, customPath string) (sourceText string, err error) {
	username := account.Credentials.GetUserName()

	// We are ignoring errors here, and checking lower down if the creds are missing
	credentialsPath, _ := buildClusterCredentialsPath(username, name, customPath)
	creds, _ := common.NewCredentialsBundle(credentialsPath)

	shellScriptPath := getCredentialFilePath(credentialsPath, shell)
	_, err = os.Stat(shellScriptPath)

	// Re-download the credentials bundle, if files are missing or the credentials are invalid
	if os.IsNotExist(err) || creds.Verify() != nil {
		credentialsPath, err = client.DownloadClusterCredentials(account, name, customPath)
		if err != nil {
			return "", err
		}
	}

	sourceText = sourceHelpString(shellScriptPath, name, shell)
	return sourceText, nil
}

// ListClusters retrieves all clusters
func (client *Client) ListClusters(account *Account) ([]common.Cluster, error) {
	svc, err := client.buildContainerService(account)
	if err != nil {
		return nil, err
	}

	return svc.ListClusters()
}

// GetCluster retrieves a cluster
func (client *Client) GetCluster(account *Account, name string, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.GetCluster(name)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// GrowCluster adds nodes to a cluster
func (client *Client) GrowCluster(account *Account, name string, nodes int, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.GrowCluster(name, nodes)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (client *Client) RebuildCluster(account *Account, name string, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.RebuildCluster(name)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// SetAutoScale adds nodes to a cluster
func (client *Client) SetAutoScale(account *Account, name string, value bool) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	return svc.SetAutoScale(name, value)
}

// DeleteCluster deletes a cluster
func (client *Client) DeleteCluster(account *Account, name string) (common.Cluster, error) {
	var cluster common.Cluster

	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.DeleteCluster(name)
	if err == nil {
		err = client.DeleteClusterCredentials(account, name, "")
	}

	return cluster, err
}

// DeleteClusterCredentials removes a cluster's downloaded credentials
func (client *Client) DeleteClusterCredentials(account *Account, name string, customPath string) error {
	username := account.Credentials.GetUserName()
	p, err := buildClusterCredentialsPath(username, name, customPath)
	if err != nil {
		common.Log.WriteWarning("Unable to locate carina config path, not deleteing credentials on disk.")
		return err
	}

	p = filepath.Clean(p)
	if p == "" || p == "." || p == "/" {
		return errors.New("Path to cluster is empty, the current directory, or a root path, not deleting")
	}

	_, statErr := os.Stat(p)
	if os.IsNotExist(statErr) {
		// Assume credentials were never on disk
		return nil
	}

	// If the path exists but not the actual credentials, inform user
	_, statErr = os.Stat(filepath.Join(p, "ca.pem"))
	if os.IsNotExist(statErr) {
		return errors.New("Path to cluster credentials exists but not the ca.pem, not deleting")
	}

	err = os.RemoveAll(p)
	if err != nil {
		return errors.Wrap(err, "Unable to delete the credentials on disk")
	}

	return nil
}
