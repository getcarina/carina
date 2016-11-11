package magnum

import (
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/getcarina/carina/common"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/pkg/errors"
)

// Account is a set of authentication credentials accepted by OpenStack Identity (keystone) v2 and v3
type Account struct {
	AuthEndpoint     string
	EndpointOverride string
	UserName         string
	Password         string
	Project          string
	Domain           string
	Region           string
	token            string
	endpoint         string
}

// GetID returns a unique id for the account, e.g. private-[authendpoint hash]-[username]
func (account *Account) GetID() string {
	hash := sha1.Sum([]byte(account.AuthEndpoint))
	return fmt.Sprintf("private-%x-%s", hash[:4], account.UserName)
}

// GetClusterPrefix returns a unique string to identity the account's clusters, e.g. private-[endpoint hash]-[username]
func (account *Account) GetClusterPrefix() (string, error) {
	endpoint := account.getEndpoint()
	if endpoint == "" {
		return "", errors.New("Cannot call account.GetClusterPrefix before authenticating and setting account.Endpoint")
	}

	hash := sha1.Sum([]byte(endpoint))
	return fmt.Sprintf("private-%x-%s", hash[:4], account.UserName), nil
}

func (account *Account) getEndpoint() string {
	if account.EndpointOverride != "" {
		return account.EndpointOverride
	}
	return account.endpoint
}

// Authenticate creates an authenticated client, ready to use to communicate with the OpenStack Magnum API
func (account *Account) Authenticate() (*gophercloud.ServiceClient, error) {
	var magnumClient *gophercloud.ServiceClient

	testAuth := func() error {
		req, err := http.NewRequest("HEAD", account.AuthEndpoint+"/auth/tokens", nil)
		if err != nil {
			return err
		}
		req.Header.Add("X-Auth-Token", account.token)
		req.Header.Add("X-Subject-Token", account.token)
		resp, err := common.NewHTTPClient().Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Cached token is invalid")
		}

		return nil
	}

	authOptions := &gophercloud.AuthOptions{
		IdentityEndpoint: account.AuthEndpoint,
		Username:         account.UserName,
		Password:         account.Password,
		TenantName:       account.Project,
		DomainName:       account.Domain,
		TokenID:          account.token,
	}

	if account.token != "" && account.endpoint != "" {
		common.Log.WriteDebug("[magnum] Attempting to authenticate with a cached token for %s", account.endpoint)
		if testAuth() == nil {
			identity, err := openstack.NewClient(account.AuthEndpoint)
			if err != nil {
				return nil, errors.Wrap(err, "[magnum] Unable to create a new OpenStack Identity client")
			}

			identity.TokenID = account.token
			identity.ReauthFunc = reauthenticate(identity, authOptions)
			identity.UserAgent.Prepend(common.BuildUserAgent())
			identity.HTTPClient = *common.NewHTTPClient()
			identity.EndpointLocator = func(opts gophercloud.EndpointOpts) (string, error) {
				// Skip the service catalog and use the cached endpoint
				return account.endpoint, nil
			}

			magnumClient, err = openstack.NewContainerOrchestrationV1(identity, gophercloud.EndpointOpts{Region: account.Region})
			if err != nil {
				return nil, errors.Wrap(err, "[magnum] Unable to create a Magnum client")
			}
		}

		// Otherwise we fall through and authenticate with the password
		common.Log.WriteDebug("[magnum] Discarding expired cached token and endpoint")
		account.token = ""
		account.endpoint = ""
	} else {
		common.Log.WriteDebug("[magnum] Attempting to authenticate with a password")
		identity, err := openstack.AuthenticatedClient(*authOptions)
		if err != nil {
			return nil, errors.Wrap(err, "[magnum] Authentication failed")
		}
		magnumClient, err = openstack.NewContainerOrchestrationV1(identity, gophercloud.EndpointOpts{Region: account.Region})
		if err != nil {
			return nil, errors.Wrap(err, "[magnum] Unable to create a Magnum client")
		}
	}
	common.Log.WriteDebug("[magnum] Authentication sucessful")

	// Apply our HTTP client customizations
	magnumClient.UserAgent.Prepend(common.BuildUserAgent())
	magnumClient.HTTPClient = *common.NewHTTPClient()

	// Cache data looked up from the service catalog
	account.token = magnumClient.TokenID
	account.endpoint = magnumClient.Endpoint // don't cache the overridden endpoint!

	// Override the endpoint from the service catalog
	magnumClient.Endpoint = account.getEndpoint()

	return magnumClient, nil
}

func reauthenticate(identity *gophercloud.ProviderClient, authOptions *gophercloud.AuthOptions) func() error {
	return func() error {
		return openstack.Authenticate(identity, *authOptions)
	}
}

// BuildCache builds the set of data to cache
func (account *Account) BuildCache() map[string]string {
	return map[string]string{
		"endpoint": account.endpoint,
		"token":    account.token,
	}
}

// ApplyCache applies a set of cached data
func (account *Account) ApplyCache(c map[string]string) {
	account.endpoint = c["endpoint"]
	account.token = c["token"]
}
