package makecoe

import "github.com/getcarina/libcarina"

// Cluster represents a cluster on make-coe
type Cluster libcarina.Cluster

// GetName returns the cluster name
func (cluster Cluster) GetName() string {
	return cluster.Name
}

// GetCOE returns the container orchestration engine used by the cluster
func (cluster Cluster) GetCOE() string {
	return cluster.COE
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster Cluster) GetFlavor() string {
	return "siracha" // TODO: find this data!
}

// GetNodes returns the number of nodes in the cluster
func (cluster Cluster) GetNodes() int {
	return cluster.Nodes
}

// GetStatus returns the status of the cluster
func (cluster Cluster) GetStatus() string {
	return cluster.Status
}
