package makecoe

import (
	"strconv"

	"github.com/getcarina/libcarina"
)

// Cluster represents a cluster on make-coe
type Cluster struct {
	*libcarina.Cluster
}

func newCluster() *Cluster {
	return &Cluster{Cluster: &libcarina.Cluster{}}
}

// GetID returns the cluster identifier
func (cluster *Cluster) GetID() string {
	return cluster.ID
}

// GetName returns the cluster name
func (cluster *Cluster) GetName() string {
	return cluster.Name
}

// GetType returns the container orchestration engine used by the cluster
func (cluster *Cluster) GetType() string {
	return cluster.Type.COE
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster *Cluster) GetFlavor() string {
	return ""
}

// GetNodes returns the number of nodes in the cluster
func (cluster *Cluster) GetNodes() string {
	return strconv.Itoa(cluster.Nodes)
}

// GetStatus returns the status of the cluster
func (cluster *Cluster) GetStatus() string {
	return cluster.Status
}

// GetStatusDetails is not supported
func (cluster *Cluster) GetStatusDetails() string {
	return ""
}
