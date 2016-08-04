package client

// Account contains the data required to communicate with a Carina API instance
type Account interface {
	Cacheable

	// GetID returns a unique string to identity for the account
	GetID() string
}
