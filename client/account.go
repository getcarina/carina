package client

import "github.com/getcarina/carina/common"

// Account contains the data required to communicate with a Carina API instance
type Account interface {
	Cacheable

	// GetID returns a unique string to identity to the account's credentials
	GetID() string

	// GetClusterPrefix returns a unique string to identity the account's clusters
	GetClusterPrefix() (string, error)

	// NewClusterService create the appropriate ClusterService for the account
	NewClusterService() common.ClusterService
}
