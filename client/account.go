package client

type Account struct {
	CloudType     string
	Credentials   Credentials
}

type Credentials interface {
	GetUserName() string
}