package client

import (
	"crypto/sha1"
	"fmt"
)

// Account contains the data required to communicate with a Carina API instance
type Account struct {
	CloudType   string
	Credentials UserCredentials
}

// UserCredentials is a set of user credentials to authenticate to a Carina API instance
type UserCredentials interface {
	// GetEndpoint returns the API endpoint
	GetEndpoint() string

	// GetUserName returns the username
	GetUserName() string

	// GetToken returns the API authentication token
	GetToken() string
}

// GetTag returns a unique tag for the account, e.g. public-8e3d76d3
func (account *Account) GetTag() string {
	endpoint := account.Credentials.GetEndpoint()
	hash := sha1.Sum([]byte(endpoint))
	return fmt.Sprintf("%s-%x", account.CloudType, hash[:4])
}
