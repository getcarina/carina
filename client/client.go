package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/magnum"
	"github.com/getcarina/carina/make-coe"
	"github.com/getcarina/carina/makeswarm"
	"github.com/pkg/errors"
)

// Client is the multi-cloud Carina client, which coorindates communication with all Carina-esque clouds
type Client struct {
	Cache *Cache
	Error error
}

// CarinaHomeDirEnvVar is the environment variable name for carina data, config, etc.
const CarinaHomeDirEnvVar = "CARINA_HOME"

// CloudMakeSwarm is the v1 Carina (make-swarm) cloud type
const CloudMakeSwarm = "make-swarm"

// CloudMakeCOE is the v2 Carina (make-coe) cloud type
const CloudMakeCOE = "public"

// CloudMagnum is the Rackspace Private Cloud Magnum cloud type
const CloudMagnum = "private"

// NewClient builds a new Carina client
func NewClient(cacheEnabled bool) *Client {
	client := &Client{}
	client.initCache(cacheEnabled)
	return client
}

func (client *Client) initCache(cacheEnabled bool) {
	disableCache := func(err error) {
		common.Log.WriteWarning("Unable to initialize cache. Starting fresh!")
		common.Log.WriteWarning(err.Error())
		client.Cache = &Cache{}
		client.Error = CacheUnavailableError{cause: err}
	}

	if !cacheEnabled {
		common.Log.WriteDebug("Cache disabled")
		client.Cache = &Cache{}
		return
	}

	bd, err := GetCredentialsDir()
	if err != nil {
		disableCache(err)
		return
	}

	err = os.MkdirAll(bd, 0777)
	if err != nil {
		disableCache(errors.Wrap(err, "Unable to create cache directory"))
		return
	}

	path, err := defaultCacheFilename()
	if err != nil {
		disableCache(err)
		return
	}

	client.Cache = newCache(path)
	err = client.Cache.load()
	if err != nil {
		disableCache(err)
	}
}

func (client *Client) buildContainerService(account Account) (common.ClusterService, error) {
	client.Cache.apply(account)

	switch a := account.(type) {
	case *makecoe.Account:
		return &makecoe.MakeCOE{Account: a}, nil
	case *makeswarm.Account:
		return &makeswarm.MakeSwarm{Account: a}, nil
	case *magnum.Account:
		return &magnum.Magnum{Account: a}, nil
	default:
		return nil, fmt.Errorf("Invalid account type: %T", a)
	}
}

// GetQuotas retrieves the quotas set for the account
func (client *Client) GetQuotas(account Account) (common.Quotas, error) {
	var quotas common.Quotas

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return quotas, err
	}

	return svc.GetQuotas()
}

// CreateCluster creates a new cluster and prints the cluster information
func (client *Client) CreateCluster(account Account, name string, template string, nodes int, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.CreateCluster(name, template, nodes)

	if waitUntilActive && err == nil {
		cluster, err = svc.WaitUntilClusterIsActive(cluster)
	}

	return cluster, err
}

// DownloadClusterCredentials downloads the TLS certificates and configuration scripts for a cluster
func (client *Client) DownloadClusterCredentials(account Account, name string, customPath string) (credentialsPath string, err error) {
	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return "", err
	}

	creds, err := svc.GetClusterCredentials(name)
	if err != nil {
		return "", err
	}

	credentialsPath, err = buildClusterCredentialsPath(account, name, customPath)
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
func (client *Client) GetSourceCommand(account Account, shell string, name string, customPath string) (sourceText string, err error) {
	// We are ignoring errors here, and checking lower down if the creds are missing
	credentialsPath, _ := buildClusterCredentialsPath(account, name, customPath)
	creds := common.LoadCredentialsBundle(credentialsPath)

	// Re-download the credentials bundle, if the credentials are invalid
	err = creds.Verify()
	if err != nil {
		common.Log.Debug(err)
		common.Log.Debugln("Re-downloading credentials due to missing or invalid credentials bundle.")

		credentialsPath, err = client.DownloadClusterCredentials(account, name, customPath)
		if err != nil {
			return "", err
		}
	}

	shellScriptPath, err := getCredentialScriptPath(credentialsPath, shell)
	if err != nil {
		return "", err
	}

	sourceText = sourceHelpString(shellScriptPath, name, shell)
	return sourceText, nil
}

// ListClusters retrieves all clusters
func (client *Client) ListClusters(account Account) ([]common.Cluster, error) {
	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return nil, err
	}

	return svc.ListClusters()
}

// GetCluster retrieves a cluster
func (client *Client) GetCluster(account Account, name string, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.GetCluster(name)

	if waitUntilActive && err == nil {
		cluster, err = svc.WaitUntilClusterIsActive(cluster)
	}

	return cluster, err
}

// GrowCluster adds nodes to a cluster
func (client *Client) GrowCluster(account Account, name string, nodes int, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.GrowCluster(name, nodes)

	if waitUntilActive && err == nil {
		cluster, err = svc.WaitUntilClusterIsActive(cluster)
	}

	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (client *Client) RebuildCluster(account Account, name string, waitUntilActive bool) (common.Cluster, error) {
	var cluster common.Cluster

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	cluster, err = svc.RebuildCluster(name)

	if waitUntilActive && err == nil {
		cluster, err = svc.WaitUntilClusterIsActive(cluster)
	}

	return cluster, err
}

// SetAutoScale adds nodes to a cluster
func (client *Client) SetAutoScale(account Account, name string, value bool) (common.Cluster, error) {
	var cluster common.Cluster

	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return cluster, err
	}

	return svc.SetAutoScale(name, value)
}

// DeleteCluster deletes a cluster
func (client *Client) DeleteCluster(account Account, name string, waitUntilDeleted bool) error {
	defer client.Cache.SaveAccount(account)
	svc, err := client.buildContainerService(account)
	if err != nil {
		return err
	}

	cluster, err := svc.DeleteCluster(name)

	if waitUntilDeleted && err == nil {
		err = svc.WaitUntilClusterIsDeleted(cluster)
	}

	if err == nil {
		err = client.DeleteClusterCredentials(account, name, "")
	}

	return err
}

// DeleteClusterCredentials removes a cluster's downloaded credentials
func (client *Client) DeleteClusterCredentials(account Account, name string, customPath string) error {
	p, err := buildClusterCredentialsPath(account, name, customPath)
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
