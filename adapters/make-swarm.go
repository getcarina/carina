package adapters

import (
	"fmt"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"strconv"
	"text/tabwriter"
	"time"
)

// MakeSwarm is an adapter between the cli and Carina (make-swarm)
type MakeSwarm struct {
	Credentials UserCredentials
	Output      *tabwriter.Writer
}

const httpTimeout = time.Second * 15

// LoadCredentials accepts credentials collected by the cli
func (carina *MakeSwarm) LoadCredentials(credentials UserCredentials) {
	carina.Credentials = credentials
}

func (carina *MakeSwarm) authenticate() (*libcarina.ClusterClient, error) {
	fmt.Println("[DEBUG][make-swarm] Authenticating...")
	carinaClient, err := libcarina.NewClusterClient(carina.Credentials.Endpoint, carina.Credentials.UserName, carina.Credentials.Secret)
	if err == nil {
		carinaClient.Client.Timeout = httpTimeout
	}
	if err != nil {
		err = errors.Wrap(err, "[make-swarm] Authentication failed")
	}
	return carinaClient, err
}

// ListClusters prints out a list of the user's clusters to the console
func (carina *MakeSwarm) ListClusters() error {
	carinaClient, err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Println("[DEBUG][make-swarm] Listing clusters")
	clusterList, err := carinaClient.List()
	if err != nil {
		return errors.Wrap(err, "Unable to list clusters")
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
func (carina *MakeSwarm) ShowCluster(name string) error {
	carinaClient, err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Showing cluster: %s\n", name)
	cluster, err := carinaClient.Get(name)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to show cluster (%s)", name))
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
	carinaClient, err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Deleting cluster: %s\n", name)
	cluster, err := carinaClient.Delete(name)
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
	carinaClient, err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Growing cluster (%s) by %d nodes\n", name, nodes)
	cluster, err := carinaClient.Grow(name, nodes)
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
	carinaClient, err := carina.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][make-swarm] Changing the autoscale setting on the cluster (%s) to %t\n", name, value)
	cluster, err := carinaClient.SetAutoScale(name, value)
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

func (carina *MakeSwarm) writeCluster(cluster *libcarina.Cluster) error {
	fields := []string{
		cluster.ClusterName,
		cluster.Flavor,
		strconv.FormatInt(cluster.Nodes.Int64(), 10),
		strconv.FormatBool(cluster.AutoScale),
		cluster.Status,
	}
	return writeRow(carina.Output, fields)
}

func (carina *MakeSwarm) writeClusterHeader() error {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"AutoScale",
		"Status",
	}
	return writeRow(carina.Output, headerFields)
}
