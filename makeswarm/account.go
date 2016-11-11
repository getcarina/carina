package makeswarm

import (
	"fmt"
	"net/http"

	"github.com/getcarina/carina/common"
	libcarina "github.com/getcarina/libmakeswarm"
	"github.com/pkg/errors"
	"github.com/rackspace/gophercloud/rackspace"
)

// Account is a set of authentication credentials accepted by Rackspace Identity
type Account struct {
	UserName string
	APIKey   string
	token    string
	endpoint string
}

// GetID returns a unique id for the account, e.g. public-[username]
func (account *Account) GetID() string {
	return fmt.Sprintf("public-%s", account.UserName)
}

// GetClusterPrefix returns a unique string to identity the account's clusters, e.g. makeswarm-[username]
func (account *Account) GetClusterPrefix() (string, error) {
	return fmt.Sprintf("makeswarm-%s", account.UserName), nil
}

// Authenticate creates an authenticated client, ready to use to communicate with the Carina API
func (account *Account) Authenticate() (*libcarina.ClusterClient, error) {
	var carinaClient *libcarina.ClusterClient

	testAuth := func() error {
		req, err := http.NewRequest("HEAD", rackspace.RackspaceUSIdentity+"tokens/"+account.token, nil)
		if err != nil {
			return err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-Auth-Token", account.token)
		req.Header.Add("User-Agent", common.BuildUserAgent())

		resp, err := common.NewHTTPClient().Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Cached token is invalid")
		}

		return nil
	}

	if account.token != "" && account.endpoint != "" {
		common.Log.WriteDebug("[make-swarm] Attempting to authenticate with a cached token")
		if testAuth() == nil {
			common.Log.WriteDebug("[make-swarm] Authentication sucessful")
			carinaClient = &libcarina.ClusterClient{
				Client:    common.NewHTTPClient(),
				Username:  account.UserName,
				Token:     account.token,
				Endpoint:  libcarina.BetaEndpoint,
				UserAgent: common.BuildUserAgent(),
			}
			return carinaClient, nil
		}

		// Otherwise we fall through and authenticate with the apikey
		common.Log.WriteDebug("[make-swarm] Discarding expired cached token")
		account.token = ""
	}

	common.Log.WriteDebug("[make-swarm] Attempting to authenticate with an apikey")
	carinaClient, err := libcarina.NewClusterClient(libcarina.BetaEndpoint, account.UserName, account.APIKey)
	if err != nil {
		return nil, errors.Wrap(err, "[make-swarm] Authentication failed")
	}
	common.Log.WriteDebug("[make-swarm] Authentication sucessful")

	carinaClient.Client = common.NewHTTPClient()
	carinaClient.UserAgent = common.BuildUserAgent()
	account.token = carinaClient.Token

	return carinaClient, nil
}

// BuildCache builds the set of data to cache
func (account *Account) BuildCache() map[string]string {
	return map[string]string{
		"token":    account.token,
		"endpoint": account.endpoint,
	}
}

// ApplyCache applies a set of cached data
func (account *Account) ApplyCache(c map[string]string) {
	account.token = c["token"]
	account.endpoint = c["endpoint"]
}
