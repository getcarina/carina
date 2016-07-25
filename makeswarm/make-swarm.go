package makeswarm

import (
	"fmt"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"net/http"
	"text/tabwriter"
	"time"
)

// MakeSwarm is an adapter between the cli and Carina (make-swarm)
type MakeSwarm struct {
	client      *libcarina.ClusterClient
	Credentials Credentials
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
		fmt.Println("[DEBUG][make-swarm] Authenticating")
		carinaClient, err := libcarina.NewClusterClient(carina.Credentials.Endpoint, carina.Credentials.UserName, carina.Credentials.APIKey)
		if err != nil {
			return errors.Wrap(err, "[make-swarm] Authentication failed")

		}

		carinaClient.Client.Timeout = httpTimeout
		carina.client = carinaClient
	}
	return nil
}

// CreateCluster creates a new cluster and prints the cluster information
func (carina *MakeSwarm) CreateCluster(name string, nodes int, waitUntilActive bool) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][make-swarm] Creating %d-node cluster (%s)\n", nodes, name)
	options := Cluster{
		ClusterName: name,
		Nodes:       libcarina.Number(nodes),
		AutoScale:   false, // Not exposing this since we are removing autoscale in make-coe
	}
	cluster, err := carina.client.Create(options)
	if err != nil {
		err = errors.Wrap(err, "[make-swarm] Unable to create the cluster")
	} else if waitUntilActive {
		time.Sleep(initialClusterWaitTime)
		cluster, err = carina.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// ListClusters prints out a list of the user's clusters to the console
func (carina *MakeSwarm) ListClusters() ([]*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Println("[DEBUG][make-swarm] Listing clusters")
	clusters, err := carina.client.List()
	if err != nil {
		err = errors.Wrap(err, "[make-swarm] Unable to list clusters")
	}

	return clusters, err
}

// RebuildCluster destroys and recreates the cluster
func (carina *MakeSwarm) RebuildCluster(name string, waitUntilActive bool) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][make-swarm] Rebuilding cluster (%s)\n", name)
	cluster, err := carina.client.Rebuild(name)
	if err != nil {
		err = errors.Wrap(err, "[make-swarm] Unable to rebuild the cluster")
	} else if waitUntilActive {
		time.Sleep(initialClusterWaitTime)
		cluster, err = carina.WaitUntilClusterIsActive(name)
	}

	return cluster, err
}

// ShowCluster prints out a cluster's information to the console
func (carina *MakeSwarm) GetCluster(name string, waitUntilActive bool) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][make-swarm] Retrieving cluster (%s)\n", name)
	var cluster *Cluster
	if waitUntilActive {
		cluster, err = carina.WaitUntilClusterIsActive(name)
	} else {
		cluster, err = carina.client.Get(name)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to retrieve cluster (%s)", name))
		}
	}
	return cluster, err
}

// DeleteCluster permanently deletes a cluster
func (carina *MakeSwarm) DeleteCluster(name string) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][make-swarm] Deleting cluster (%s)\n", name)
	cluster, err := carina.client.Delete(name)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to delete cluster (%s)", name))
	}
	return cluster, err
}

// GrowCluster adds nodes to a cluster
func (carina *MakeSwarm) GrowCluster(name string, nodes int) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][make-swarm] Growing cluster (%s) by %d nodes\n", name, nodes)
	cluster, err := carina.client.Grow(name, nodes)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to grow cluster (%s)", name))
	}

	return cluster, err
}

// SetAutoScale enables or disables autoscaling on a cluster
func (carina *MakeSwarm) SetAutoScale(name string, value bool) error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Changing the autoscale setting on the cluster (%s) to %t\n", name, value)
	cluster, err := carina.client.SetAutoScale(name, value)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to change the cluster's autoscale setting (%s)", name))
	}

	return cluster, err
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (carina *MakeSwarm) WaitUntilClusterIsActive(name string) (*Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	checkCluster := func() (*Cluster, error) {
		carina.client.Client = &http.Client{Timeout: httpTimeout}
		cluster, err := carina.client.Get(name)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to retrieve cluster (%s)", name))
		}
		return cluster, err
	}

	var cluster *Cluster
	for {
		cluster, err = checkCluster()
		if err != nil {
			return nil, err
		}

		// Transitions past point of "new" or "building" are assumed to be active states
		if cluster.Status != StatusNew && cluster.Status != StatusBuilding && cluster.Status != StatusRebuilding {
			return cluster, nil
		}
		time.Sleep(clusterPollingInterval)
	}
}
