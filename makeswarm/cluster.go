package makeswarm

import "github.com/getcarina/libmakeswarm"

// Cluster represents a cluster on make-swarm
type Cluster libcarina.Cluster

// GetName returns the cluster name
func (cluster Cluster) GetName() string {
	return cluster.ClusterName
}

// GetCOE returns the container orchestration engine used by the cluster
func (cluster Cluster) GetCOE() string {
	return "swarm"
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster Cluster) GetFlavor() string {
	return cluster.Flavor
}

// GetNodes returns the number of nodes in the cluster
func (cluster Cluster) GetNodes() int {
	return cluster.Nodes.Int()
}

// GetStatus returns the status of the cluster
func (cluster Cluster) GetStatus() string {
	return cluster.Status
}
