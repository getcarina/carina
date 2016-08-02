package client

import (
	"crypto/sha1"
	"fmt"
)

type Account struct {
	CloudType   string
	Credentials UserCredentials
}

type UserCredentials interface {
	GetEndpoint() string
	GetUserName() string
	GetToken() string
}

// GetTag returns a unique tag for the account, e.g. public-8e3d76d3
func (account *Account) GetTag() string {
	endpoint := account.Credentials.GetEndpoint()
	hash := sha1.Sum([]byte(endpoint))
	return fmt.Sprintf("%s-%x", account.CloudType, hash[:4])
}
