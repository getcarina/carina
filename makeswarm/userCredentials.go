package makeswarm

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/libcarina"
	"github.com/pkg/errors"
	"net/http"
)

// Credentials is a set of authentication credentials accepted by Rackspace Identity
type UserCredentials struct {
	Endpoint string
	UserName string
	APIKey   string
	Token    string
}

// Authenticate creates an authenticated client, ready to use to communicate with the Carina API
func (credentials *UserCredentials) Authenticate() (*libcarina.ClusterClient, error) {
	var carinaClient *libcarina.ClusterClient

	testAuth := func(c *libcarina.ClusterClient) error {
		req, err := http.NewRequest("HEAD", c.Endpoint+"/clusters/"+c.Username, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "getcarina/carina dummy request")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-Auth-Token", c.Token)
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Unable to auth on %s", "/clusters"+c.Username)
		}

		return nil
	}

	if credentials.Token != "" {
		common.Log.WriteDebug("[make-swarm] Attempting to authenticate with a cached token")
		carinaClient = &libcarina.ClusterClient{
			Client:   &http.Client{Timeout: httpTimeout},
			Username: credentials.UserName,
			Token:    credentials.Token,
			Endpoint: credentials.Endpoint,
		}

		if testAuth(carinaClient) == nil {
			return carinaClient, nil
		}

		// Otherwise we fall through and authenticate with the apikey
		common.Log.WriteDebug("[make-swarm] Discarding expired cached token")
		credentials.Token = ""
	}

	common.Log.WriteDebug("[make-swarm] Attempting to authenticate with an apikey")
	carinaClient, err := libcarina.NewClusterClient(credentials.Endpoint, credentials.UserName, credentials.APIKey)
	if err != nil {
		return nil, errors.Wrap(err, "[make-swarm] Authentication failed")
	}

	common.Log.WriteDebug("[make-swarm] Authentication sucessful")
	credentials.Token = carinaClient.Token
	carinaClient.Client.Timeout = httpTimeout

	return carinaClient, nil
}

// GetEndpoint returns the API endpoint
func (credentials *UserCredentials) GetEndpoint() string {
	return credentials.Endpoint
}

// GetUserName returns the username
func (credentials *UserCredentials) GetUserName() string {
	return credentials.UserName
}

// GetToken returns the API authentication token
func (credentials *UserCredentials) GetToken() string {
	return credentials.Token
}
