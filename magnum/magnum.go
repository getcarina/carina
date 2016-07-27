package magnum

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
	"strings"
	"text/tabwriter"
	"time"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	client      *gophercloud.ServiceClient
	Credentials MagnumCredentials
	Output      *tabwriter.Writer
}

const clusterPollingInterval = 10 * time.Second

func (magnum *Magnum) authenticate() error {
	if magnum.client == nil {
		common.Log.WriteDebug("[magnum] Authenticating")
		auth := gophercloud.AuthOptions{
			IdentityEndpoint: magnum.Credentials.Endpoint,
			Username:         magnum.Credentials.UserName,
			Password:         magnum.Credentials.Password,
			TenantName:       magnum.Credentials.Project,
			DomainName:       magnum.Credentials.Domain,
		}
		identity, err := openstack.AuthenticatedClient(auth)
		if err != nil {
			return errors.Wrap(err, "[magnum] Authentication failed")
		}
		magnum.client, err = openstack.NewContainerOrchestrationV1(identity, gophercloud.EndpointOpts{Region: magnum.Credentials.Region})
		if err != nil {
			return errors.Wrap(err, "[magnum] Unable to create a Magnum client")
		}
	}
	return nil
}

// GetQuotas retrieves the quotas set for the account
func (magnum *Magnum) GetQuotas() (common.Quotas, error) {
	return MagnumQuotas{}, errors.New("Not implemented yet")
}

// CreateCluster creates a new cluster and prints the cluster information
func (magnum *Magnum) CreateCluster(name string, nodes int) (common.Cluster, error) {
	return MagnumCluster{}, errors.New("Not implemented yet")
}

// GetClusterCredentials retrieves the TLS certificates and configuration scripts for a cluster
func (magnum *Magnum) GetClusterCredentials(name string) (common.CredentialsBundle, error) {
	return common.CredentialsBundle{}, errors.New("Not implemented yet")
}

// ListClusters prints out a list of the user's clusters to the console
func (magnum *Magnum) ListClusters() ([]common.Cluster, error) {
	err := magnum.authenticate()
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
			cluster := MagnumCluster(result)
			clusters = append(clusters, cluster)
		}
		return true, nil
	})

	return clusters, err
}

// GetCluster prints out a cluster's information to the console
func (magnum *Magnum) GetCluster(name string) (common.Cluster, error) {
	var cluster MagnumCluster

	err := magnum.authenticate()
	if err != nil {
		return cluster, err
	}

	common.Log.WriteDebug("[magnum] Retrieving bay (%s)", name)
	result, err := bays.Get(magnum.client, name).Extract()
	if err != nil {
		return cluster, errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve bay (%s)", name))
	}
	cluster = MagnumCluster(*result)

	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (magnum *Magnum) RebuildCluster(name string) (common.Cluster, error) {
	return MagnumCluster{}, errors.New("Not implemented yet")
}

// DeleteCluster permanently deletes a cluster
func (magnum *Magnum) DeleteCluster(name string) (common.Cluster, error) {
	return MagnumCluster{}, errors.New("Not implemented yet")
}

// GrowCluster adds nodes to a cluster
func (magnum *Magnum) GrowCluster(name string, nodes int) (common.Cluster, error) {
	return MagnumCluster{}, errors.New("Not implemented yet")
}

// SetAutoScale enables or disables autoscaling on a cluster
func (magnum *Magnum) SetAutoScale(name string, value bool) (common.Cluster, error) {
	return MagnumCluster{}, errors.New("Magnum does not support autoscaling.")
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
