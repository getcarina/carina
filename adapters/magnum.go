package adapters

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

type Magnum struct {
	Credentials UserCredentials
}

func (magnum *Magnum) LoadCredentials(credentials UserCredentials) error {
	magnum.Credentials = credentials
	return nil
}

func (magnum *Magnum) ListClusters() error {
	fmt.Println("[DEBUG] Listing Magnum bays...")

	auth := gophercloud.AuthOptions{
		IdentityEndpoint: magnum.Credentials.Endpoint,
		Username:         magnum.Credentials.UserName,
		Password:         magnum.Credentials.Secret,
		TenantName:       magnum.Credentials.Project,
		DomainName:       magnum.Credentials.Domain,
	}
	identityService, authErr := openstack.AuthenticatedClient(auth)
	if authErr != nil {
		return authErr
	}
	fmt.Println(identityService.TokenID)

	return nil
}
