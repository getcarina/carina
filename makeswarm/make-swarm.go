package makeswarm

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"text/tabwriter"
	"time"
)

// MakeSwarm is an adapter between the cli and Carina (make-swarm)
type MakeSwarm struct {
	client      *libcarina.ClusterClient
	Credentials UserCredentials
	Output      *tabwriter.Writer
}

// StatusNew is the status of a new, inactive cluster
const StatusNew = "new"

// StatusBuilding is the status of a cluster that is currently being built
const StatusBuilding = "building"

// StatusRebuilding is the status of a cluster that is currently rebuilding
const StatusRebuilding = "rebuilding-swarm"

const httpTimeout = 15 * time.Second
const initialClusterWaitTime = 1 * time.Minute
const clusterPollingInterval = 10 * time.Second

func (carina *MakeSwarm) authenticate() error {
	if carina.client == nil {
		common.Log.WriteDebug("[make-swarm] Authenticating")
		carinaClient, err := libcarina.NewClusterClient(carina.Credentials.Endpoint, carina.Credentials.UserName, carina.Credentials.APIKey)
		if err != nil {
			return errors.Wrap(err, "[make-swarm] Authentication failed")

		}

		carinaClient.Client.Timeout = httpTimeout
		carina.client = carinaClient
	}
	return nil
}

func (carina *MakeSwarm) GetQuotas() (common.Quotas, error) {
	var quotas common.Quotas

	err := carina.authenticate()
	if err != nil {
		return quotas, err
	}

	common.Log.WriteDebug("[make-swarm] Retrieving account quotas")
	carina.client.GetQuotas()

	return quotas, err
}

// CreateCluster creates a new cluster and prints the cluster information
func (carina *MakeSwarm) CreateCluster(name string, nodes int) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Creating %d-node cluster (%s)", nodes, name)
	options := libcarina.Cluster{
		ClusterName: name,
		Nodes:       libcarina.Number(nodes),
		AutoScale:   false, // Not exposing this since we are removing autoscale in make-coe
	}
	result, err := carina.client.Create(options)
	if err != nil {
		return cluster, errors.Wrap(err, "[make-swarm] Unable to create the cluster")
	}
	cluster = CarinaCluster(*result)

	return cluster, err
}

// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster
func (carina *MakeSwarm) GetClusterCredentials(name string) (common.CredentialsBundle, error) {
	return common.CredentialsBundle{}, errors.New("Not implemented yet")
}

// ListClusters prints out a list of the user's clusters to the console
func (carina *MakeSwarm) ListClusters() ([]common.Cluster, error) {
	var clusters []common.Cluster

	err := carina.authenticate()
	if err != nil {
		return clusters, err
	}

	common.Log.WriteDebug("[make-swarm] Listing clusters")
	results, err := carina.client.List()
	if err != nil {
		return clusters, errors.Wrap(err, "[make-swarm] Unable to list clusters")
	}

	for _, result := range results {
		clusters = append(clusters, CarinaCluster(result))
	}

	return clusters, err
}

// RebuildCluster destroys and recreates the cluster
func (carina *MakeSwarm) RebuildCluster(name string) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Rebuilding cluster (%s)", name)
	result, err := carina.client.Rebuild(name)
	cluster = CarinaCluster(*result)

	if err != nil {
		return cluster, errors.Wrap(err, "[make-swarm] Unable to rebuild the cluster")
	}

	return cluster, nil
}

// ShowCluster prints out a cluster's information to the console
func (carina *MakeSwarm) GetCluster(name string) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Retrieving cluster (%s)", name)
	result, err := carina.client.Get(name)
	if err != nil {
		return cluster, errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to retrieve cluster (%s)", name))
	}
	cluster = CarinaCluster(*result)

	return cluster, nil
}

// DeleteCluster permanently deletes a cluster
func (carina *MakeSwarm) DeleteCluster(name string) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Deleting cluster (%s)", name)
	result, err := carina.client.Delete(name)
	if err != nil {
		return cluster, errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to delete cluster (%s)", name))
	}
	cluster = CarinaCluster(*result)

	return cluster, nil
}

// GrowCluster adds nodes to a cluster
func (carina *MakeSwarm) GrowCluster(name string, nodes int) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Growing cluster (%s) by %d nodes", name, nodes)
	result, err := carina.client.Grow(name, nodes)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to grow cluster (%s)", name))
	}
	cluster = CarinaCluster(*result)

	return cluster, nil
}

// SetAutoScale enables or disables autoscaling on a cluster
func (carina *MakeSwarm) SetAutoScale(name string, value bool) (common.Cluster, error) {
	var cluster CarinaCluster

	err := carina.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[make-swarm] Changing the autoscale setting on the cluster (%s) to %t", name, value)
	result, err := carina.client.SetAutoScale(name, value)
	if err != nil {
		return cluster, errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to change the cluster's autoscale setting (%s)", name))
	}
	cluster = CarinaCluster(*result)

	return cluster, nil
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (carina *MakeSwarm) WaitUntilClusterIsActive(name string) (common.Cluster, error) {
	for {
		cluster, err := carina.GetCluster(name)
		if err != nil {
			return cluster, err
		}

		// Transitions past point of "new" or "building" are assumed to be active states
		status := cluster.GetStatus()
		if status != StatusNew && status != StatusBuilding && status != StatusRebuilding {
			return cluster, nil
		}
		time.Sleep(clusterPollingInterval)
	}
}
