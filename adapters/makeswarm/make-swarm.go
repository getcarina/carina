package makeswarm

import (
	"fmt"
	"github.com/getcarina/carina/adapters"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"text/tabwriter"
	"time"
)

// MakeSwarm is an adapter between the cli and Carina (make-swarm)
type MakeSwarm struct {
	client      *libcarina.ClusterClient
	Credentials Credentials
	Output      *tabwriter.Writer
}

// Credentials is a set of authentication credentials accepted by Rackspace Identity
type Credentials struct {
	Endpoint        string
	UserName        string
	APIKey          string
	Project         string
	Token           string
	TokenExpiration string
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
func (carina *MakeSwarm) CreateCluster(name string, nodes int, waitUntilActive bool) error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Creating %d-node cluster (%s)\n", nodes, name)
	options := libcarina.Cluster{
		ClusterName: name,
		Nodes:       libcarina.Number(nodes),
		AutoScale:   false, // Not exposing this since we are removing autoscale in make-coe
	}
	cluster, err := carina.client.Create(options)
	if err != nil {
		return errors.Wrap(err, "[make-swarm] Unable to list clusters")
	}

	if waitUntilActive {
		time.Sleep(initialClusterWaitTime)
		cluster, err = carina.waitUntilClusterIsActive(name)
		if err != nil {
			return err
		}
	}

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	err = carina.writeCluster(cluster)
	return err
}

// ListClusters prints out a list of the user's clusters to the console
func (carina *MakeSwarm) ListClusters() error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Println("[DEBUG][make-swarm] Listing clusters")
	clusterList, err := carina.client.List()
	if err != nil {
		return errors.Wrap(err, "[make-swarm] Unable to list clusters")
	}

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	for _, cluster := range clusterList {
		err = carina.writeCluster(&cluster)
		if err != nil {
			return err
		}
	}

	return err
}

// ShowCluster prints out a cluster's information to the console
func (carina *MakeSwarm) ShowCluster(name string, waitUntilActive bool) error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Retrieving cluster (%s)\n", name)
	var cluster *libcarina.Cluster
	if waitUntilActive {
		cluster, err = carina.waitUntilClusterIsActive(name)
	} else {
		cluster, err = carina.client.Get(name)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to retrieve cluster (%s)", name))
		}
	}
	if err != nil {
		return err
	}

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	err = carina.writeCluster(cluster)
	return err
}

// DeleteCluster permanently deletes a cluster
func (carina *MakeSwarm) DeleteCluster(name string) error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Deleting cluster (%s)\n", name)
	cluster, err := carina.client.Delete(name)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to delete cluster (%s)", name))
	}

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	err = carina.writeCluster(cluster)
	return err
}

// GrowCluster adds nodes to a cluster
func (carina *MakeSwarm) GrowCluster(name string, nodes int) error {
	err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Growing cluster (%s) by %d nodes\n", name, nodes)
	cluster, err := carina.client.Grow(name, nodes)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to grow cluster (%s)", name))
	}

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	err = carina.writeCluster(cluster)
	return err
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

	err = carina.writeClusterHeader()
	if err != nil {
		return err
	}

	err = carina.writeCluster(cluster)
	return err
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (carina *MakeSwarm) waitUntilClusterIsActive(name string) (*libcarina.Cluster, error) {
	err := carina.authenticate()
	if err != nil {
		return nil, err
	}

	checkCluster := func() (*libcarina.Cluster, error) {
		carina.client.Client = &http.Client{Timeout: httpTimeout}
		cluster, err := carina.client.Get(name)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[make-swarm] Unable to retrieve cluster (%s)", name))
		}
		return cluster, err
	}

	var cluster *libcarina.Cluster
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

func (carina *MakeSwarm) writeCluster(cluster *libcarina.Cluster) error {
	fields := []string{
		cluster.ClusterName,
		cluster.Flavor,
		strconv.FormatInt(cluster.Nodes.Int64(), 10),
		strconv.FormatBool(cluster.AutoScale),
		cluster.Status,
	}
	return adapters.WriteRow(carina.Output, fields)
}

func (carina *MakeSwarm) writeClusterHeader() error {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"AutoScale",
		"Status",
	}
	return adapters.WriteRow(carina.Output, headerFields)
}
