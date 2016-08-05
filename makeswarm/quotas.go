package makeswarm

import "github.com/getcarina/libcarina"

// CarinaQuotas contains the quota information for a CarinaAccount
type CarinaQuotas libcarina.Quotas

// GetMaxClusters returns the maximum number of clusters allowed on the account
func (quotas CarinaQuotas) GetMaxClusters() int {
	return quotas.MaxClusters.Int()
}

// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
func (quotas CarinaQuotas) GetMaxNodesPerCluster() int {
	return quotas.MaxNodesPerCluster.Int()
}
