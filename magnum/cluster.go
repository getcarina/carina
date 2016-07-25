package magnum

import "github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"

type Cluster bays.Bay

func (cluster *Cluster) GetName() string {
	return cluster.Name
}

func (cluster *Cluster) GetFlavor() string {
	return "" // TODO lookup the baymodel
}

func (cluster *Cluster) GetNodes() int {
	return cluster.Nodes
}

func (cluster *Cluster) GetStatus() string {
	return cluster.Status
}