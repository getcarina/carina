package magnum

// Credentials is a set of authentication credentials accepted by OpenStack Identity (keystone) v2 and v3
type MagnumCredentials struct {
	Endpoint string
	UserName string
	Password string
	Project  string
	Domain   string
	Region   string
	Token    string
}

func (credentials MagnumCredentials) GetEndpoint() string {
	return credentials.Endpoint
}

func (credentials MagnumCredentials) GetUserName() string {
	return credentials.UserName
}

func (credentials MagnumCredentials) GetToken() string {
	return credentials.Token
}
