package makeswarm

import (
	"strconv"

	libcarina "github.com/getcarina/libmakeswarm"
)

// Cluster represents a cluster on make-swarm
type Cluster libcarina.Cluster

// GetID returns the cluster identifier
func (cluster Cluster) GetID() string {
	return cluster.ClusterName
}

// GetName returns the cluster name
func (cluster Cluster) GetName() string {
	return cluster.ClusterName
}

// GetType returns the container orchestration engine used by the cluster
func (cluster Cluster) GetType() string {
	return "swarm"
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster Cluster) GetFlavor() string {
	return cluster.Flavor
}

// GetNodes returns the number of nodes in the cluster
func (cluster Cluster) GetNodes() string {
	return strconv.Itoa(cluster.Nodes.Int())
}

// GetStatus returns the status of the cluster
func (cluster Cluster) GetStatus() string {
	return cluster.Status
}

// GetStatusDetails is not supported
func (cluster Cluster) GetStatusDetails() string {
	return ""
}
