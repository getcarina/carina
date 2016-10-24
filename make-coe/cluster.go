package makecoe

import (
	"strconv"

	"github.com/getcarina/carina/common"
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

// GetTemplate returns the template used to create the cluster
func (cluster *Cluster) GetTemplate() common.ClusterTemplate {
	return &ClusterTemplate{ClusterType: cluster.Type}
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
