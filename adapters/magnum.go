package adapters

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/gophercloud/gophercloud/openstack/containerorchestration/v1/bays"
	"text/tabwriter"
	"strconv"
)

type Magnum struct {
	Credentials UserCredentials
	Output *tabwriter.Writer
}

func (magnum *Magnum) LoadCredentials(credentials UserCredentials) error {
	magnum.Credentials = credentials
	return nil
}

func (magnum *Magnum) authenticate() (*gophercloud.ServiceClient, error) {
	fmt.Println("[DEBUG][Magnum] Authenticating...")
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
	return openstack.NewContainerOrchestrationV1(identity, gophercloud.EndpointOpts{ Region: magnum.Credentials.Region})
}

func (magnum *Magnum) ListClusters() error {
	magnumService, err := magnum.authenticate()
	if err != nil {
		return errors.Wrap(err, "[Magnum] Authentication failed")
	}

	fmt.Println("[DEBUG][Magnum] Listing clusters...")
	results := bays.List(magnumService, bays.ListOpts{})
	if results.Err != nil {
		return errors.Wrap(results.Err, "[Magnum] Unable to list clusters")
	}

	err = magnum.writeClusterHeader()
	if err != nil {
		return err
	}

	err = results.EachPage(func(page pagination.Page) (bool, error) {
		clusters, err := bays.ExtractBays(page)
		if err != nil {
			return false, errors.Wrap(err, "[Magnum]Unable to read the Magnum clusters from the results page")
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
