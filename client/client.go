package client

import (
	"os"

	"github.com/getcarina/carina/magnum"
	"github.com/getcarina/carina/makeswarm"
	"github.com/pkg/errors"
	"fmt"
	"path/filepath"
)

type Client struct {
	cacheEnabled bool
	Cache *Cache
	Error error
}

// CarinaHomeDirEnvVar is the environment variable name for carina data, config, etc.
const CarinaHomeDirEnvVar = "CARINA_HOME"

const CloudMakeSwarm = "make-swarm"
const CloudMakeCOE = "make-coe"
const CloudMagnum = "magnum"

func NewClient(cacheEnabled bool) *Client {
	client := &Client{cacheEnabled: cacheEnabled}
	err := client.initCache()
	if err != nil {
		client.cacheEnabled = false
		client.Error = CacheUnavailableError { cause: err }
	}
	return client, err
}

func (client *Client) initCache() error {
	if client.cacheEnabled {
		bd, err := CarinaCredentialsBaseDir()
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

func (client *Client) buildContainerService(account *Account) clusterService {
	switch account.CloudType {
	case CloudMakeSwarm:
		return &makeswarm.MakeSwarm{Credentials: account.Credentials}, nil
	case CloudMagnum:
		return &magnum.Magnum{Credentials: account.Credentials}, nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", account.CloudType)
	}
}

// CreateCluster creates a new cluster and prints the cluster information
func (client *Client) CreateCluster(account *Account, name string, nodes int, waitUntilActive bool) (*Cluster, error) {
	svc := client.buildContainerService(account)
	cluster, err := svc.CreateCluster(name, nodes)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// ListClusters retrieves all clusters
func (client *Client) ListClusters(account *Account) ([]*Cluster, error) {
	adapter := client.buildContainerService(account.CloudType)
	return adapter.ListClusters()
}

// GetCluster retrieves a cluster
func (client *Client) GetCluster(account *Account, name string, waitUntilActive bool) (*Cluster, error) {
	svc := client.buildContainerService(account.CloudType)
	cluster, err := svc.GetCluster(name)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}
	return cluster, err
}

// GrowCluster adds nodes to a cluster
func (client *Client) GrowCluster(account *Account, name string, nodes int, waitUntilActive bool) (*Cluster, error) {
	svc := client.buildContainerService(account)
	cluster, err := svc.GrowCluster(name, nodes)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (client *Client) RebuildCluster(account *Account, name string, waitUntilActive bool) (*Cluster, error) {
	svc := client.buildContainerService(account)
	cluster, err := svc.RebuildCluster(name)
	if waitUntilActive && err != nil {
		cluster, err = svc.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// SetAutoScale adds nodes to a cluster
func (client *Client) SetAutoScale(account *Account, name string, value bool) (*Cluster, error) {
	svc := client.buildContainerService(account)
	return svc.SetAutoScale(name, value)
}

// DeleteCluster deletes a cluster
func (client *Client) DeleteCluster(account *Account, name string) (*Cluster, error) {
	svc := client.buildContainerService(account)
	cluster, err := svc.DeleteCluster(name)
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
		fmt.Fprintf(os.Stderr, "[WARN] Unable to locate carina config path, not deleteing credentials on disk\n")
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