package makeswarm

import "github.com/getcarina/libcarina"

type CarinaQuotas libcarina.Quotas

func (quotas CarinaQuotas) GetMaxClusters() int {
	return quotas.MaxClusters.Int()
}

func (quotas CarinaQuotas) GetMaxNodesPerCluster() int {
	return quotas.MaxNodesPerCluster.Int()
}
