package makeswarm

import (
	"strconv"

	"github.com/getcarina/carina/common"
	libcarina "github.com/getcarina/libmakeswarm"
)

// Cluster represents a cluster on make-swarm
type Cluster struct {
	*libcarina.Cluster
	Template *ClusterTemplate
}

func newCluster() *Cluster {
	return &Cluster{
		Cluster:  &libcarina.Cluster{},
		Template: &ClusterTemplate{},
	}
}

// GetID returns the cluster identifier
func (cluster *Cluster) GetID() string {
	return cluster.ClusterName
}

// GetName returns the cluster name
func (cluster *Cluster) GetName() string {
	return cluster.ClusterName
}

// GetTemplate returns the template used to create the cluster
func (cluster *Cluster) GetTemplate() common.ClusterTemplate {
	return cluster.Template
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster *Cluster) GetFlavor() string {
	return cluster.Flavor
}

// GetNodes returns the number of nodes in the cluster
func (cluster *Cluster) GetNodes() string {
	return strconv.Itoa(cluster.Nodes.Int())
}

// GetStatus returns the status of the cluster
func (cluster *Cluster) GetStatus() string {
	return cluster.Status
}

// GetStatusDetails is not supported
func (cluster *Cluster) GetStatusDetails() string {
	return ""
}
