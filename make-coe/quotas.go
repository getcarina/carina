package makecoe

// Quotas contains the quota information for a CarinaAccount
type Quotas struct{}

// GetMaxClusters returns the maximum number of clusters allowed on the account
func (quotas *Quotas) GetMaxClusters() int {
	return 3
}

// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
func (quotas *Quotas) GetMaxNodesPerCluster() int {
	return 1
}
