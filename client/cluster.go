package client

type Cluster interface {
	GetName() string
	GetFlavor() string
	GetNodes() int
	GetStatus() string
}
