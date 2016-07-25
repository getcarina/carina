package makeswarm

import "github.com/getcarina/libcarina"

type CarinaCluster libcarina.Cluster

func (cluster CarinaCluster) GetName() string {
	return cluster.ClusterName
}

func (cluster CarinaCluster) GetFlavor() string {
	return cluster.Flavor
}

func (cluster CarinaCluster) GetNodes() int {
	return cluster.Nodes.Int()
}

func (cluster CarinaCluster) GetStatus() string {
	return cluster.Status
}
