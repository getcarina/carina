package magnum

// Credentials is a set of authentication credentials accepted by OpenStack Identity (keystone) v2 and v3
type Credentials struct {
	Endpoint        string
	UserName        string
	Password        string
	Project         string
	Domain          string
	Region          string
	Token           string
	TokenExpiration string
}

func (credentials *Credentials) GetUserName() string {
	return credentials.UserName
}
