package makecoe

import (
	"crypto/sha1"
	"fmt"

	"regexp"

	"strings"

	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
)

// Account is a set of authentication credentials accepted by Rackspace Identity
type Account struct {
	// Optional custom endpoint specified by the user
	EndpointOverride string
	UserName         string
	APIKey           string
	Region           string

	// Testing only, not used by the cli
	AuthEndpointOverride string

	// The endpoint from the service catalog
	endpoint string
	token    string
}

// GetID returns a unique id for the account, e.g. public-[username]
func (account *Account) GetID() string {
	return fmt.Sprintf("public-%s", account.UserName)
}

// GetClusterPrefix returns a unique string to identity the account's clusters, e.g. public-[region]-[username]
func (account *Account) GetClusterPrefix() (string, error) {
	endpoint := account.getEndpoint()
	if endpoint == "" {
		return "", errors.New("Cannot call account.GetClusterPrefix before authenticating and setting account.Endpoint")
	}

	region := account.getEndpointRegion()
	if region == "" {
		// This is a not a production API endpoint, just use a hash for local dev testing to avoid collisions
		hash := sha1.Sum([]byte(endpoint))
		return fmt.Sprintf("public-%x-%s", hash[:4], account.UserName), nil
	}

	return fmt.Sprintf("public-%s-%s", strings.ToLower(region), account.UserName), nil
}

func (account *Account) getEndpoint() string {
	if account.EndpointOverride != "" {
		return account.EndpointOverride
	}
	return account.endpoint
}

func (account *Account) getEndpointRegion() string {
	re := regexp.MustCompile(`https://api\.([^.]*)\.getcarina\.com`)
	match := re.FindStringSubmatch(account.getEndpoint())
	if len(match) < 2 {
		return ""
	}

	return match[1]
}

// Authenticate creates an authenticated client, ready to use to communicate with the Carina API
func (account *Account) Authenticate() (*libcarina.CarinaClient, error) {
	if account.token != "" && account.endpoint != "" {
		common.Log.WriteDebug("[make-coe] Attempting to authenticate with a cached token, falling back to the username and apikey if necessary")
	} else {
		common.Log.WriteDebug("[make-coe] Attempting to authenticate with a username and apikey")
	}
	carinaClient, err := libcarina.NewClient(account.UserName, account.APIKey, account.Region, account.AuthEndpointOverride, account.token, account.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "[make-coe] Authentication failed")
	}
	common.Log.WriteDebug("[make-coe] Authentication sucessful")

	// Apply our http client customizations
	carinaClient.Client = common.NewHTTPClient()
	carinaClient.UserAgent = common.BuildUserAgent()

	// Cache data looked up from the service catalog
	account.token = carinaClient.Token
	account.endpoint = carinaClient.Endpoint // don't cache the overridden endpoint!

	// Override the endpoint from the service catalog
	carinaClient.Endpoint = account.getEndpoint()

	common.Log.WriteDebug("[make-coe] Checking server API version")
	metadata, err := carinaClient.GetAPIMetadata()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse API version response")
	}
	if !metadata.IsSupportedVersion() {
		min, max := metadata.GetSupportedVersionRange()
		return nil, fmt.Errorf("Unable to communicate with the Carina API because the client is out-of-date. The client supports ~%s while the server supports %s-%s. Update the carina client to the latest version. See https://getcarina.com/docs/tutorials/carina-cli#update for instructions.", libcarina.SupportedAPIVersion, min, max)
	}

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
