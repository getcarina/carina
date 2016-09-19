package makecoe

import "github.com/getcarina/libcarina"

// Quotas contains the quota information for a CarinaAccount
type Quotas libcarina.Quotas

// GetMaxClusters returns the maximum number of clusters allowed on the account
func (quotas *Quotas) GetMaxClusters() int {
	return quotas.MaxClusters
}

// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
func (quotas *Quotas) GetMaxNodesPerCluster() int {
	return quotas.MaxNodesPerCluster
}
