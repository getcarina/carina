package client

type Account struct {
	CloudType   string
	Credentials UserCredentials
}

type UserCredentials interface {
	GetUserName() string
}
