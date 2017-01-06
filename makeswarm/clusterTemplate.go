package makeswarm

// ClusterTemplate represents a cluster template for makeswarm
type ClusterTemplate struct {
}

// GetName returns the unique template name
func (template *ClusterTemplate) GetName() string {
	return ""
}

// GetCOE returns the container orchestration engine used by the cluster
func (template *ClusterTemplate) GetCOE() string {
	return "swarm"
}

// GetHostType returns the underlying type of the host nodes, such as lxc or vm
func (template *ClusterTemplate) GetHostType() string {
	return "lxc"
}
