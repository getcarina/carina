package makecoe

import (
	"crypto/sha1"
	"fmt"

	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
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
	return libcarina.CarinaEndpoint
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
	common.Log.WriteDebug("[make-coe] Attempting to authenticate")
	carinaClient, err := libcarina.NewClient(account.getEndpoint(), account.UserName, account.APIKey, account.Token)
	if err != nil {
		return nil, errors.Wrap(err, "[make-coe] Authentication failed")
	}
	common.Log.WriteDebug("[make-coe] Authentication sucessful")

	// Apply our http client customizations
	carinaClient.Client = common.NewHTTPClient()
	carinaClient.UserAgent = common.BuildUserAgent()

	// Cache the token
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
