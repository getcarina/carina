package magnum

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/baymodels"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
)

// Cluster is a Magnum cluster
type Cluster struct {
	*bays.Bay
	Template *baymodels.BayModel
}

func newCluster() *Cluster {
	return &Cluster{Bay: &bays.Bay{}}
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
	return cluster.Template.COE
}

// GetFlavor returns the flavor of the nodes in the cluster
func (cluster *Cluster) GetFlavor() string {
	return cluster.Template.FlavorID
}

// GetNodes returns the number of nodes in the cluster
func (cluster *Cluster) GetNodes() string {
	return fmt.Sprintf("%d/%d", cluster.Masters, cluster.Nodes)
}

// GetStatus returns the status of the cluster
func (cluster *Cluster) GetStatus() string {
	return cluster.Status
}

// GetStatusDetails returns additional information about the cluster's status.
// For example, why the cluster is in a failed state.
func (cluster *Cluster) GetStatusDetails() string {
	if strings.HasSuffix(cluster.Status, "FAILED") {
		return cluster.StatusReason
	}

	return ""
}
