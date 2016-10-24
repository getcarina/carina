package magnum

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/pem"

	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/baymodels"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/certificates"
	coe "github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/common"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	client        *gophercloud.ServiceClient
	bayModelCache map[string]*baymodels.BayModel
	Account       *Account
}

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
	return nil, errors.New("Not implemented yet")
}

// CreateCluster creates a new cluster and prints the cluster information
func (magnum *Magnum) CreateCluster(name string, template string, nodes int) (common.Cluster, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Creating %d-node %s cluster (%s)", nodes, template, name)

	bayModel, err := magnum.lookupBayModelByName(template)
	if err != nil {
		return nil, err
	}

	options := bays.CreateOpts{
		Name:       name,
		BayModelID: bayModel.ID,
		Nodes:      nodes,
	}
	result := bays.Create(magnum.client, options)
	if result.Err != nil {
		return nil, errors.Wrap(result.Err, "[magnum] Unable to create the cluster")
	}

	bay, err := result.Extract()
	cluster := &Cluster{Bay: bay, Template: bayModel}

	return cluster, err
}

// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster by its id or name (if unique)
func (magnum *Magnum) GetClusterCredentials(token string) (*libcarina.CredentialsBundle, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Generating credentials bundle for cluster (%s)", token)

	result, err := certificates.CreateCredentialsBundle(magnum.client, token)
	if err != nil {
		return nil, errors.Wrap(err, "[magnum] Unable to generate credentials bundle")
	}

	creds := libcarina.NewCredentialsBundle()
	creds.Files["ca.pem"] = pem.EncodeToMemory(&result.CACertificate)
	creds.Files["key.pem"] = pem.EncodeToMemory(&result.PrivateKey)
	creds.Files["cert.pem"] = pem.EncodeToMemory(&result.Certificate)
	for filename, script := range result.Scripts {
		creds.Files[filename] = script
	}

	return creds, nil
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

		for i := range results {
			cluster, err := magnum.newCluster(&results[i])
			if err != nil {
				return false, err
			}
			clusters = append(clusters, cluster)
		}
		return true, nil
	})

	return clusters, err
}

// ListClusterTemplates retrieves available templates for creating a new cluster
func (magnum *Magnum) ListClusterTemplates() ([]common.ClusterTemplate, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	results, err := magnum.listBayModels()
	if err != nil {
		return nil, err
	}

	var templates []common.ClusterTemplate
	for _, result := range results {
		template := &ClusterTemplate{BayModel: result}
		templates = append(templates, template)
	}
	return templates, err
}

// GetCluster prints out a cluster's information to the console by its id or name (if unique)
func (magnum *Magnum) GetCluster(token string) (common.Cluster, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Retrieving bay (%s)", token)
	result, err := bays.Get(magnum.client, token).Extract()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve bay (%s)", token))
	}

	cluster, err := magnum.newCluster(result)
	return cluster, err
}

// RebuildCluster destroys and recreates the cluster by its id or name (if unique)
func (magnum *Magnum) RebuildCluster(token string) (common.Cluster, error) {
	return nil, errors.New("Not implemented yet")
}

// DeleteCluster permanently deletes a cluster by its id or name (if unique)
func (magnum *Magnum) DeleteCluster(token string) (common.Cluster, error) {
	err := magnum.init()
	if err != nil {
		return nil, err
	}

	common.Log.WriteDebug("[magnum] Deleting cluster (%s)", token)
	result := bays.Delete(magnum.client, token)
	if result.Err != nil {
		return nil, errors.Wrap(result.Err, fmt.Sprintf("[magnum] Unable to delete cluster (%s)", token))
	}

	cluster, err := magnum.waitForTaskInitiated(token, "DELETE")
	if err != nil {
		err = errors.Cause(err)

		// Gracefully handle a 404 Not Found when the cluster is deleted quickly
		if httpErr, ok := err.(*coe.ErrorResponse); ok {
			if httpErr.Actual == http.StatusNotFound {
				cluster = newCluster()
				cluster.Status = "DELETE_COMPLETE"
				return cluster, nil
			}
		} else {
			return nil, err
		}
	}

	return cluster, err
}

// GrowCluster adds nodes to a cluster by its id or name (if unique)
func (magnum *Magnum) GrowCluster(token string, nodes int) (common.Cluster, error) {
	return nil, errors.New("Not implemented yet")
}

// SetAutoScale is not supported
func (magnum *Magnum) SetAutoScale(token string, value bool) (common.Cluster, error) {
	return nil, errors.New("Magnum does not support autoscaling.")
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (magnum *Magnum) WaitUntilClusterIsActive(cluster common.Cluster) (common.Cluster, error) {
	isDone := func(cluster common.Cluster) bool {
		status := strings.ToLower(cluster.GetStatus())
		return !strings.HasSuffix(status, "in_progress")
	}

	if isDone(cluster) {
		return cluster, nil
	}

	pollingInterval := 10 * time.Second
	for {
		cluster, err := magnum.GetCluster(cluster.GetID())
		if err != nil {
			return cluster, err
		}

		if isDone(cluster) {
			return cluster, nil
		}

		common.Log.WriteDebug("[magnum] Waiting until cluster (%s) is active, currently in %s", cluster.GetName(), cluster.GetStatus())
		time.Sleep(pollingInterval)
	}
}

// WaitUntilClusterIsDeleted polls the cluster status until either the cluster is gone or an error state is hit
func (magnum *Magnum) WaitUntilClusterIsDeleted(cluster common.Cluster) error {
	isDone := func(cluster common.Cluster) bool {
		status := strings.ToUpper(cluster.GetStatus())
		return status == "DELETE_COMPLETE"
	}

	if isDone(cluster) {
		return nil
	}

	pollingInterval := 5 * time.Second
	for {
		cluster, err := magnum.GetCluster(cluster.GetID())

		if err != nil {
			err = errors.Cause(err)

			// Gracefully handle a 404 Not Found when the cluster is deleted quickly
			if httpErr, ok := err.(*coe.ErrorResponse); ok {
				if httpErr.Actual == http.StatusNotFound {
					return nil
				}
			}

			return err
		}

		if isDone(cluster) {
			return nil
		}

		common.Log.WriteDebug("[magnum] Waiting until cluster (%s) is deleted, currently in %s", cluster.GetName(), cluster.GetStatus())
		time.Sleep(pollingInterval)
	}
}

// waitForClusterStatus waits for a cluster to reach a particular group of states, e.g. delete will
// wait for DELETE_IN_PROGRESS, DELETE_FAILED or DELETE_COMPLETE. This is necessary as the Magnum API
// returns immediately and updates the status later
func (magnum *Magnum) waitForTaskInitiated(token string, task string) (*Cluster, error) {
	task = strings.ToLower(task)

	pollingInterval := 1 * time.Second
	for {
		result, err := magnum.GetCluster(token)
		cluster, _ := result.(*Cluster)
		if err != nil {
			return cluster, err
		}

		status := strings.ToLower(cluster.Status)
		if strings.HasPrefix(status, task) {
			return cluster, nil
		}

		common.Log.WriteDebug("[magnum] Waiting for %s_* currently in %s", task, status)
		time.Sleep(pollingInterval)
	}
}

func (magnum *Magnum) newCluster(bay *bays.Bay) (*Cluster, error) {
	cluster := &Cluster{Bay: bay}
	baymodel, err := magnum.lookupBayModelByID(bay.BayModelID)
	cluster.Template = baymodel
	return cluster, err
}

func (magnum *Magnum) listBayModels() ([]*baymodels.BayModel, error) {
	common.Log.WriteDebug("[magnum] Listing baymodels")
	pager := baymodels.List(magnum.client, baymodels.ListOpts{})
	if pager.Err != nil {
		return nil, errors.Wrap(pager.Err, "[magnum] Unabe to list baymodels")
	}

	var bayModels []*baymodels.BayModel
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		results, err := baymodels.ExtractBayModels(page)
		if err != nil {
			return false, errors.Wrap(err, "[magnum] Unable to read the Magnum baymodels from the results page")
		}

		for i := range results {
			bayModels = append(bayModels, &results[i])
		}
		return true, nil
	})

	return bayModels, err
}

func (magnum *Magnum) getBayModelCache() (map[string]*baymodels.BayModel, error) {
	if magnum.bayModelCache == nil {
		bayModels, err := magnum.listBayModels()
		if err != nil {
			return nil, err
		}

		magnum.bayModelCache = make(map[string]*baymodels.BayModel)
		for _, bayModel := range bayModels {
			magnum.bayModelCache[bayModel.ID] = bayModel
		}
	}

	return magnum.bayModelCache, nil
}

func (magnum *Magnum) lookupBayModelByID(bayModelID string) (*baymodels.BayModel, error) {
	cache, err := magnum.getBayModelCache()
	if err != nil {
		return nil, err
	}

	bayModel, exists := cache[bayModelID]
	if !exists {
		return nil, fmt.Errorf("Could not find baymodel with id %s", bayModelID)
	}
	return bayModel, nil
}

func (magnum *Magnum) lookupBayModelByName(name string) (*baymodels.BayModel, error) {
	cache, err := magnum.getBayModelCache()
	if err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	var bayModel *baymodels.BayModel
	for _, m := range cache {
		if strings.ToLower(m.Name) == name {
			bayModel = m
			break
		}
	}

	if bayModel == nil {
		return nil, fmt.Errorf("Could not find baymodel named %s", name)
	}

	return bayModel, nil
}
