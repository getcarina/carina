package adapters

// Maps between a container service API and the command line client
type Adapter interface {
	LoadCredentials(credentials UserCredentials) error
	ListClusters() error
}

// The credentials supplied by the user to the command line client
type UserCredentials struct {
	Endpoint        string
	UserName        string
	Secret          string
	Token           string
	TokenExpiration string
}
