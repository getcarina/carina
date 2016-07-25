package common

// ClusterService is a common interface over multiple cluster APIs (magnum, make-swarm and make-coe)
type ClusterService interface {
	// GetQuotas retrieves the quotas set for the account
	GetQuotas() (Quotas, error)

	// CreateCluster creates a new cluster
	CreateCluster(name string, nodes int) (Cluster, error)

	// ListClusters retrieves all clusters
	ListClusters() ([]Cluster, error)

	// GetCluster retrieves a cluster
	GetCluster(name string) (Cluster, error)

	// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster
	GetClusterCredentials(name string) (CredentialsBundle, error)

	// RebuildCluster destroys and recreates the cluster
	RebuildCluster(name string) (Cluster, error)

	// DeleteCluster permanently deletes a cluster
	DeleteCluster(name string) (Cluster, error)

	// GrowCluster adds nodes to a cluster
	GrowCluster(name string, nodes int) (Cluster, error)

	// SetAutoScale enables or disables autoscaling on a cluster
	SetAutoScale(name string, value bool) (Cluster, error)

	// WaitUntilClusterIsActive polls the cluster status until either an active or error state is hit
	WaitUntilClusterIsActive(name string) (Cluster, error)
}

type Cluster interface {
	GetName() string
	GetFlavor() string
	GetNodes() int
	GetStatus() string
}

type Quotas interface {
	GetMaxClusters() int
	GetMaxNodesPerCluster() int
}
