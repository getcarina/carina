package common

import (
	"fmt"

	"github.com/getcarina/libcarina"
)

// ClusterService is a common interface over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type ClusterService interface {
	// GetQuotas retrieves the quotas set for the account
	GetQuotas() (Quotas, error)

	// CreateCluster creates a new cluster
	CreateCluster(name string, template string, nodes int) (Cluster, error)

	// ListClusters retrieves all clusters
	ListClusters() ([]Cluster, error)

	// ListClusterTemplates retrieves available templates for creating a new cluster
	ListClusterTemplates() ([]ClusterTemplate, error)

	// GetCluster retrieves a cluster by its id or name (if unique)
	GetCluster(token string) (Cluster, error)

	// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster by its id or name (if unique)
	GetClusterCredentials(token string) (*libcarina.CredentialsBundle, error)

	// ResizeCluster resizes the cluster to the specified number of nodes
	ResizeCluster(token string, nodes int) (Cluster, error)

	// RebuildCluster destroys and recreates the cluster by its id or name (if unique)
	RebuildCluster(token string) (Cluster, error)

	// DeleteCluster permanently deletes a cluster by its id or name (if unique)
	DeleteCluster(token string) (Cluster, error)

	// GrowCluster adds nodes to a cluster by its id or name (if unique)
	GrowCluster(token string, nodes int) (Cluster, error)

	// SetAutoScale enables or disables autoscaling on a cluster by its id or name (if unique)
	SetAutoScale(token string, value bool) (Cluster, error)

	// WaitUntilClusterIsActive polls the cluster status until either an active or error state is hit
	WaitUntilClusterIsActive(cluster Cluster) (Cluster, error)

	// WaitUntilClusterIsDeleted polls the cluster status until either the cluster is gone or an error state is hit
	WaitUntilClusterIsDeleted(cluster Cluster) error
}

// Cluster is a common interface for clusters over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type Cluster interface {
	// GetID returns the cluster identifier
	GetID() string

	// GetName returns the cluster name
	GetName() string

	// GetTemplate returns the template used to create the cluster
	GetTemplate() ClusterTemplate

	// GetFlavor returns the flavor of the nodes in the cluster
	GetFlavor() string

	// GetNodes returns the number of nodes in the cluster
	GetNodes() string

	// GetStatus returns the status of the cluster
	GetStatus() string

	// GetStatusDetails returns additional information about the cluster's status.
	// For example, why the cluster is in a failed state.
	GetStatusDetails() string
}

// ClusterTemplate is a common interface for templates over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type ClusterTemplate interface {
	// GetName returns the unique template name
	GetName() string

	// GetCOE returns the container orchestration engine used by the cluster
	GetCOE() string

	// GetHostType returns the underlying type of the host nodes, such as lxc or vm
	GetHostType() string
}

// Quotas is a common interface for cluster quotas over multiple container orchestration engine APIs (magnum, make-swarm and make-coe)
type Quotas interface {
	// GetMaxClusters returns the maximum number of clusters allowed on the account
	GetMaxClusters() int

	// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
	GetMaxNodesPerCluster() int
}

// MultipleMatchingTemplatesError indicates when a template search was too broad and matched multiple templates
type MultipleMatchingTemplatesError struct {
	TemplatePattern string
}

// Error returns the underlying error message
func (error MultipleMatchingTemplatesError) Error() string {
	return fmt.Sprintf("Multiple matching templates found for '%s'. Run carina templates --name %s to refine the search pattern to only match a single template.", error.TemplatePattern, error.TemplatePattern)
}
