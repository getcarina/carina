package makecoe

import (
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"github.com/rackspace/gophercloud/rackspace"
)

// Account is a set of authentication credentials accepted by Rackspace Identity
type Account struct {
	Endpoint string
	UserName string
	APIKey   string
	Token    string
}

func (account *Account) getEndpoint() string {
	if account.Endpoint != "" {
		return account.Endpoint
	}
	panic("Endpoint not specified and make-coe doesn't have a default public endpoint yet!")
}

// GetID returns a unique id for the account, e.g. public[-custom endpoint hash]-[username]
func (account *Account) GetID() string {
	if account.Endpoint == "" {
		return fmt.Sprintf("public-%s", account.UserName)
	}

	hash := sha1.Sum([]byte(account.Endpoint))
	return fmt.Sprintf("public-%x-%s", hash[:4], account.UserName)
}

// Authenticate creates an authenticated client, ready to use to communicate with the Carina API
func (account *Account) Authenticate() (*libcarina.CarinaClient, error) {
	var carinaClient *libcarina.CarinaClient

	testAuth := func() error {
		req, err := http.NewRequest("HEAD", rackspace.RackspaceUSIdentity+"tokens/"+account.Token, nil)
		if err != nil {
			return err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-Auth-Token", account.Token)
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

	if account.Token != "" {
		common.Log.WriteDebug("[make-coe] Attempting to authenticate with a cached token")
		if testAuth() == nil {
			common.Log.WriteDebug("[make-coe] Authentication sucessful")
			carinaClient = &libcarina.CarinaClient{
				Client:    common.NewHTTPClient(),
				Username:  account.UserName,
				Token:     account.Token,
				Endpoint:  account.getEndpoint(),
				UserAgent: common.BuildUserAgent(),
			}
			return carinaClient, nil
		}

		// Otherwise we fall through and authenticate with the apikey
		common.Log.WriteDebug("[make-coe] Discarding expired cached token")
		account.Token = ""
	}

	common.Log.WriteDebug("[make-coe] Attempting to authenticate with an apikey")
	carinaClient, err := libcarina.NewClient(account.getEndpoint(), account.UserName, account.APIKey)
	if err != nil {
		return nil, errors.Wrap(err, "[make-coe] Authentication failed")
	}
	common.Log.WriteDebug("[make-coe] Authentication sucessful")

	carinaClient.Client = common.NewHTTPClient()
	carinaClient.UserAgent = common.BuildUserAgent()
	account.Token = carinaClient.Token

	return carinaClient, nil
}

// BuildCache builds the set of data to cache
func (account *Account) BuildCache() map[string]string {
	c := map[string]string{"token": account.Token}
	if account.Endpoint != "" {
		c["endpoint"] = account.Endpoint
	}
	return c
}

// ApplyCache applies a set of cached data
func (account *Account) ApplyCache(c map[string]string) {
	account.Token = c["token"]

	// Don't let a cached value nuke the endpoint specified by the user
	if account.Endpoint == "" {
		account.Endpoint = c["endpoint"]
	}
}
