package magnum

import "github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"

type MagnumCluster struct {
	bays.Bay
	FlavorID string
}

func (cluster MagnumCluster) GetName() string {
	return cluster.Name
}

func (cluster MagnumCluster) GetFlavor() string {
	return cluster.FlavorID
}

func (cluster MagnumCluster) GetNodes() int {
	return cluster.Nodes
}

func (cluster MagnumCluster) GetStatus() string {
	return cluster.Status
}
