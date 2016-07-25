package magnum

type MagnumQuotas struct {
	// TODO: Implement in gophercloud and alias its type
}

func (quotas MagnumQuotas) GetMaxClusters() int {
	return 0
}

func (quotas MagnumQuotas) GetMaxNodesPerCluster() int {
	return 0
}
