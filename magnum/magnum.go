package magnum

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
	"text/tabwriter"
	"time"
	"strings"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	client      *gophercloud.ServiceClient
	Credentials Credentials
	Output      *tabwriter.Writer
}

const clusterPollingInterval = 10 * time.Second

func (magnum *Magnum) authenticate() error {
	if magnum.client == nil {
		fmt.Println("[DEBUG][magnum] Authenticating")
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

// CreateCluster creates a new cluster and prints the cluster information
func (magnum *Magnum) CreateCluster(name string, nodes int, waitUntilActive bool) (*Cluster, error) {
	return errors.New("Not implemented yet")
}

// ListClusters prints out a list of the user's clusters to the console
func (magnum *Magnum) ListClusters() ([]*Cluster, error) {
	err := magnum.authenticate()
	if err != nil {
		return nil, errors.Wrap(err, "[magnum] Authentication failed")
	}

	fmt.Println("[DEBUG][magnum] Listing clusters")
	results := bays.List(magnum.client, bays.ListOpts{})
	if results.Err != nil {
		return errors.Wrap(results.Err, "[magnum] Unable to list clusters")
	}

	var clusters []Cluster
	err = results.EachPage(func(page pagination.Page) (bool, error) {
		items, err := bays.ExtractBays(page)
		if err != nil {
			return false, errors.Wrap(err, "[magnum] Unable to read the Magnum clusters from the results page")
		}

		append(clusters, items)
		return true, nil
	})

	return clusters, err
}

// ShowCluster prints out a cluster's information to the console
func (magnum *Magnum) ShowCluster(name string, waitUntilActive bool) (*Cluster, error) {
	err := magnum.authenticate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG][magnum] Retrieving cluster (%s)\n", name)
	cluster, err := bays.Get(magnum.client, name).Extract()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve cluster (%s)", name))
	}

	return cluster, err
}

// RebuildCluster destroys and recreates the cluster
func (magnum *Magnum) RebuildCluster(name string, waitUntilActive bool) (*Cluster, error) {
	return errors.New("Not implemented yet")
}

// DeleteCluster permanently deletes a cluster
func (magnum *Magnum) DeleteCluster(name string) (*Cluster, error) {
	return errors.New("Not implemented yet")
}

// GrowCluster adds nodes to a cluster
func (magnum *Magnum) GrowCluster(name string, nodes int) (*Cluster, error) {
	return errors.New("Not implemented yet")
}

// SetAutoScale enables or disables autoscaling on a cluster
func (magnum *Magnum) SetAutoScale(name string, value bool) (*Cluster, error) {
	return errors.New("Magnum does not support autoscaling.")
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (magnum *Magnum) WaitUntilClusterIsActive(name string) (*Cluster, error) {
	err := magnum.authenticate()
	if err != nil {
		return nil, err
	}

	checkCluster := func() (*Cluster, error) {
		cluster, err := bays.Get(magnum.client, name).Extract()
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve cluster (%s)", name))
		}
		return cluster, err
	}

	var cluster *Cluster
	for {
		cluster, err = checkCluster()
		if err != nil {
			return nil, err
		}

		if !strings.HasSuffix(cluster.Status, "IN_PROGRESS") {
			return cluster, nil
		}
		time.Sleep(clusterPollingInterval)
	}
}