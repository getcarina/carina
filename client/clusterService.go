package client

// adapter maps between a container service API and the cli
type clusterService interface {
	// CreateCluster creates a new cluster
	CreateCluster(name string, nodes int) (*Cluster, error)

	// ListClusters retrieves all clusters
	ListClusters() ([]*Cluster, error)

	// ShowCluster retrieves a cluster
	GetCluster(name string) (*Cluster, error)

	// RebuildCluster destroys and recreates the cluster
	RebuildCluster(name string) (*Cluster, error)

	// DeleteCluster permanently deletes a cluster
	DeleteCluster(name string) (*Cluster, error)

	// GrowCluster adds nodes to a cluster
	GrowCluster(name string, nodes int) (*Cluster, error)

	// SetAutoScale enables or disables autoscaling on a cluster
	SetAutoScale(name string, value bool) (*Cluster, error)

	// WaitUntilClusterIsActive polls the cluster status until either an active or error state is hit
	WaitUntilClusterIsActive(name string) (*Cluster, error)
}