package makeswarm

// Credentials is a set of authentication credentials accepted by Rackspace Identity
type UserCredentials struct {
	Endpoint        string
	UserName        string
	APIKey          string
	Token           string
	TokenExpiration string
}

func (credentials UserCredentials) GetUserName() string {
	return credentials.UserName
}

func (credentials UserCredentials) GetEndpoint() string {
	return credentials.Endpoint
}
