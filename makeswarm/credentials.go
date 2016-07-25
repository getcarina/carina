package makeswarm

// Credentials is a set of authentication credentials accepted by Rackspace Identity
type Credentials struct {
	Endpoint        string
	UserName        string
	APIKey          string
	Token           string
	TokenExpiration string
}

func (credentials *Credentials) GetUserName() string {
	return credentials.UserName
}