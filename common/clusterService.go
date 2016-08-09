package common

// ClusterService is a common interface over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type ClusterService interface {
	// GetQuotas retrieves the quotas set for the account
	GetQuotas() (Quotas, error)

	// CreateCluster creates a new cluster
	CreateCluster(name string, template string, nodes int) (Cluster, error)

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

// Cluster is a common interface for clusters over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type Cluster interface {
	// GetName returns the cluster name
	GetName() string

	// GetCOE returns the container orchestration engine used by the cluster
	GetCOE() string

	// GetFlavor returns the flavor of the nodes in the cluster
	GetFlavor() string

	// GetNodes returns the number of nodes in the cluster
	GetNodes() int

	// GetStatus returns the status of the cluster
	GetStatus() string
}

// Quotas is a common interface for cluster quotas over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type Quotas interface {
	// GetMaxClusters returns the maximum number of clusters allowed on the account
	GetMaxClusters() int

	// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
	GetMaxNodesPerCluster() int
}
