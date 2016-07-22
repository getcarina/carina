package adapters

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
	"strconv"
	"text/tabwriter"
)

// Magnum is an adapter between the cli and the OpenStack COE API (Magnum)
type Magnum struct {
	Credentials UserCredentials
	Output      *tabwriter.Writer
}

// LoadCredentials accepts credentials collected by the cli
func (magnum *Magnum) LoadCredentials(credentials UserCredentials) error {
	magnum.Credentials = credentials
	return nil
}

func (magnum *Magnum) authenticate() (*gophercloud.ServiceClient, error) {
	fmt.Println("[DEBUG][magnum] Authenticating...")
	auth := gophercloud.AuthOptions{
		IdentityEndpoint: magnum.Credentials.Endpoint,
		Username:         magnum.Credentials.UserName,
		Password:         magnum.Credentials.Secret,
		TenantName:       magnum.Credentials.Project,
		DomainName:       magnum.Credentials.Domain,
	}
	identity, err := openstack.AuthenticatedClient(auth)
	if err != nil {
		return nil, err
	}
	return openstack.NewContainerOrchestrationV1(identity, gophercloud.EndpointOpts{Region: magnum.Credentials.Region})
}

// ListClusters prints out a list of the user's clusters to the console
func (magnum *Magnum) ListClusters() error {
	magnumClient, err := magnum.authenticate()
	if err != nil {
		return errors.Wrap(err, "[magnum] Authentication failed")
	}

	fmt.Println("[DEBUG][magnum] Listing clusters")
	results := bays.List(magnumClient, bays.ListOpts{})
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
func (magnum *Magnum) ShowCluster(name string) error {
	magnumClient, err := magnum.authenticate()
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG][magnum] Showing cluster: %s\n", name)
	cluster, err := bays.Get(magnumClient, name).Extract()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to show cluster (%s)", name))
	}

	err = magnum.writeClusterHeader()
	if err != nil {
		return err
	}

	err = magnum.writeCluster(cluster)
	return err
}

func (magnum *Magnum) writeCluster(cluster *bays.Bay) error {
	fields := []string{
		cluster.Name,
		"", // cluster.Flavor,
		strconv.Itoa(cluster.Nodes),
		cluster.Status,
	}
	return writeRow(magnum.Output, fields)
}

func (magnum *Magnum) writeClusterHeader() error {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"Status",
	}
	return writeRow(magnum.Output, headerFields)
}
