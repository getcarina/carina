package magnum

// Quotas contains the quota information for a MagnumAccount
type Quotas struct {
	// TODO: Implement in gophercloud and alias its type
}

// GetMaxClusters returns the maximum number of clusters allowed on the account
func (quotas *Quotas) GetMaxClusters() int {
	return 0
}

// GetMaxNodesPerCluster returns the maximum number of nodes allowed in a cluster on the account
func (quotas *Quotas) GetMaxNodesPerCluster() int {
	return 0
}
