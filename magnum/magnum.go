package magnum

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/baymodels"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	client        *gophercloud.ServiceClient
	bayModelCache map[string]baymodels.BayModel
	Account       *Account
}

const httpTimeout = 15 * time.Second

func (magnum *Magnum) init() error {
	if magnum.client == nil {
		magnumClient, err := magnum.Account.Authenticate()
		if err != nil {
			return err
		}
		magnum.client = magnumClient
	}
	return nil
}

// GetQuotas retrieves the quotas set for the account
func (magnum *Magnum) GetQuotas() (common.Quotas, error) {
	return Quotas{}, errors.New("Not implemented yet")
}

// CreateCluster creates a new cluster and prints the cluster information
func (magnum *Magnum) CreateCluster(name string, template string, nodes int) (common.Cluster, error) {
	var cluster Cluster

	err := magnum.init()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[magnum] Creating %d-node %s cluster (%s)", nodes, template, name)

	bayModel, err := magnum.lookupBayModelByName(template)
	if err != nil {
		return cluster, err
	}

	options := bays.CreateOpts{
		Name:       name,
		BayModelID: bayModel.ID,
		Nodes:      nodes,
	}
	result := bays.Create(magnum.client, options)
	if result.Err != nil {
		return cluster, errors.Wrap(result.Err, "[magnum] Unable to create the cluster")
	}

	bay, err := result.Extract()
	cluster = Cluster{Bay: *bay, Template: bayModel}

	return cluster, err
}

// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster
func (magnum *Magnum) GetClusterCredentials(name string) (common.CredentialsBundle, error) {
	return common.CredentialsBundle{}, errors.New("Not implemented yet")
}

// ListClusters prints out a list of the user's clusters to the console
func (magnum *Magnum) ListClusters() ([]common.Cluster, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Listing bays")
	pager := bays.List(magnum.client, bays.ListOpts{})
	if pager.Err != nil {
		return nil, errors.Wrap(pager.Err, "[magnum] Unable to list bays")
	}

	var clusters []common.Cluster
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		results, err := bays.ExtractBays(page)
		if err != nil {
			return false, errors.Wrap(err, "[magnum] Unable to read the Magnum bays from the results page")
		}

		for _, result := range results {
			cluster, err := magnum.newCluster(result)
			if err != nil {
				return false, err
			}
			clusters = append(clusters, cluster)
		}
		return true, nil
	})

	return clusters, err
}

// GetCluster prints out a cluster's information to the console
func (magnum *Magnum) GetCluster(name string) (common.Cluster, error) {
	var cluster Cluster

	err := magnum.init()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[magnum] Retrieving bay (%s)", name)
	result, err := bays.Get(magnum.client, name).Extract()
	if err != nil {
		common.Log.Dump(err)
		return cluster, errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve bay (%s)", name))
	}

	cluster, err = magnum.newCluster(*result)
	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (magnum *Magnum) RebuildCluster(name string) (common.Cluster, error) {
	return Cluster{}, errors.New("Not implemented yet")
}

// DeleteCluster permanently deletes a cluster
func (magnum *Magnum) DeleteCluster(name string) (common.Cluster, error) {
	var cluster Cluster

	err := magnum.init()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[magnum] Deleting cluster (%s)", name)
	result := bays.Delete(magnum.client, name)
	if result.Err != nil {
		return cluster, errors.Wrap(result.Err, fmt.Sprintf("[magnum] Unable to delete cluster (%s)", name))
	}

	return magnum.GetCluster(name)
}

// GrowCluster adds nodes to a cluster
func (magnum *Magnum) GrowCluster(name string, nodes int) (common.Cluster, error) {
	return Cluster{}, errors.New("Not implemented yet")
}

// SetAutoScale is not supported
func (magnum *Magnum) SetAutoScale(name string, value bool) (common.Cluster, error) {
	return Cluster{}, errors.New("Magnum does not support autoscaling.")
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (magnum *Magnum) WaitUntilClusterIsActive(name string) (common.Cluster, error) {
	for {
		cluster, err := magnum.GetCluster(name)
		if err != nil {
			return cluster, err
		}

		if !strings.HasSuffix(cluster.GetStatus(), "IN_PROGRESS") {
			return cluster, nil
		}
		time.Sleep(clusterPollingInterval)
	}
}

func (magnum *Magnum) newCluster(bay bays.Bay) (Cluster, error) {
	cluster := &Cluster{Bay: bay}
	baymodel, err := magnum.lookupBayModelByID(bay.BayModelID)
	cluster.Template = baymodel
	return *cluster, err
}

func (magnum *Magnum) listBayModels() ([]baymodels.BayModel, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Listing baymodels")
	pager := baymodels.List(magnum.client, baymodels.ListOpts{})
	if pager.Err != nil {
		return nil, errors.Wrap(pager.Err, "[magnum] Unabe to list baymodels")
	}

	var bayModels []baymodels.BayModel
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		results, err := baymodels.ExtractBayModels(page)
		if err != nil {
			return false, errors.Wrap(err, "[magnum] Unable to read the Magnum baymodels from the results page")
		}

		for _, result := range results {
			bayModels = append(bayModels, result)
		}
		return true, nil
	})

	return bayModels, err
}

func (magnum *Magnum) getBayModelCache() (map[string]baymodels.BayModel, error) {
	if magnum.bayModelCache == nil {
		bayModels, err := magnum.listBayModels()
		if err != nil {
			return nil, err
		}

		magnum.bayModelCache = make(map[string]baymodels.BayModel)
		for _, bayModel := range bayModels {
			magnum.bayModelCache[bayModel.ID] = bayModel
		}
	}

	return magnum.bayModelCache, nil
}

func (magnum *Magnum) lookupBayModelByID(bayModelID string) (baymodels.BayModel, error) {
	var bayModel baymodels.BayModel

	cache, err := magnum.getBayModelCache()
	if err != nil {
		return bayModel, err
	}

	bayModel, exists := cache[bayModelID]
	if !exists {
		return bayModel, fmt.Errorf("Could not find baymodel: %s", bayModelID)
	}
	return bayModel, nil
}

func (magnum *Magnum) lookupBayModelByName(name string) (baymodels.BayModel, error) {
	var bayModel baymodels.BayModel

	cache, err := magnum.getBayModelCache()
	if err != nil {
		return bayModel, err
	}

	name = strings.ToLower(name)
	for _, bayModel = range cache {
		if strings.ToLower(bayModel.Name) == name {
			break
		}
	}

	if bayModel == (baymodels.BayModel{}) {
		return bayModel, fmt.Errorf("Could not find baymodel: %s", name)
	}

	return bayModel, nil
}
