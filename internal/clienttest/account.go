package clienttest

import (
	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/stretchr/testify/mock"
)

type MockAccount struct {
	client.Account
	mock.Mock
}

func (mock *MockAccount) GetID() string {
	return "mock-user"
}

func (mock MockAccount) GetClusterPrefix() (string, error) {
	return "mock-dfw-user", nil
}

func (mock MockAccount) NewClusterService() common.ClusterService {
	args := mock.Called()
	return args.Get(0).(common.ClusterService)
}
