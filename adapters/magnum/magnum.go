package magnum

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/getcarina/carina/adapters"
	"strings"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	client      *gophercloud.ServiceClient
	Credentials Credentials
	Output      *tabwriter.Writer
}

// Credentials is a set of authentication credentials accepted by OpenStack Identity (keystone) v2 and v3
type Credentials struct {
	Endpoint        string
	UserName        string
	Password        string
	Project         string
	Domain          string
	Region          string
	Token           string
	TokenExpiration string
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
func (magnum *Magnum) CreateCluster(name string, nodes int, waitUntilActive bool) error {
	return errors.New("Not implemented yet")
}

// ListClusters prints out a list of the user's clusters to the console
func (magnum *Magnum) ListClusters() error {
	err := magnum.authenticate()
	if err != nil {
		return errors.Wrap(err, "[magnum] Authentication failed")
	}

	fmt.Println("[DEBUG][magnum] Listing clusters")
	results := bays.List(magnum.client, bays.ListOpts{})
	if results.Err != nil {
		return errors.Wrap(results.Err, "[magnum] Unable to list clusters")
	}

	err = magnum.writeClusterHeader()
	if err != nil {
		return err
	}

	err = results.EachPage(func(page pagination.Page) (bool, error) {
		clusters, err := bays.ExtractBays(page)
		if err != nil {
			return false, errors.Wrap(err, "[magnum] Unable to read the Magnum clusters from the results page")
		}

		for _, cluster := range clusters {
			err = magnum.writeCluster(&cluster)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	})

	return err
}

// ShowCluster prints out a cluster's information to the console
func (magnum *Magnum) ShowCluster(name string, waitUntilActive bool) error {
	err := magnum.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][magnum] Retrieving cluster (%s)\n", name)
	cluster, err := bays.Get(magnum.client, name).Extract()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve cluster (%s)", name))
	}

	err = magnum.writeClusterHeader()
	if err != nil {
		return err
	}

	err = magnum.writeCluster(cluster)
	return err
}

// RebuildCluster destroys and recreates the cluster
func (magnum *Magnum) RebuildCluster(name string, waitUntilActive bool) error {
	return errors.New("Not implemented yet")
}

// DeleteCluster permanently deletes a cluster
func (magnum *Magnum) DeleteCluster(name string) error {
	return errors.New("Not implemented yet")
}

// GrowCluster adds nodes to a cluster
func (magnum *Magnum) GrowCluster(name string, nodes int) error {
	return errors.New("Not implemented yet")
}

// SetAutoScale enables or disables autoscaling on a cluster
func (magnum *Magnum) SetAutoScale(name string, value bool) error {
	return errors.New("Magnum does not support autoscaling.")
}

// WaitUntilClusterIsActive waits until the prior cluster operation is completed
func (magnum *Magnum) waitUntilClusterIsActive(name string) (*bays.Bay, error) {
	err := magnum.authenticate()
	if err != nil {
		return nil, err
	}

	checkCluster := func() (*bays.Bay, error) {
		cluster, err := bays.Get(magnum.client, name).Extract()
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("[magnum] Unable to retrieve cluster (%s)", name))
		}
		return cluster, err
	}

	var cluster *bays.Bay
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

func (magnum *Magnum) writeCluster(cluster *bays.Bay) error {
	fields := []string{
		cluster.Name,
		"", // cluster.Flavor,
		strconv.Itoa(cluster.Nodes),
		cluster.Status,
	}
	return adapters.WriteRow(magnum.Output, fields)
}

func (magnum *Magnum) writeClusterHeader() error {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"Status",
	}
	return adapters.WriteRow(magnum.Output, headerFields)
}
