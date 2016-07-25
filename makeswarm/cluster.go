package makeswarm

import	"github.com/getcarina/libcarina"

type Cluster libcarina.Cluster

func (cluster *Cluster) GetName() string {
	return cluster.ClusterName
}

func (cluster *Cluster) GetFlavor() string {
	return cluster.Flavor
}

func (cluster *Cluster) GetNodes() int {
	return cluster.Nodes
}

func (cluster *Cluster) GetStatus() string {
	return cluster.Status
}