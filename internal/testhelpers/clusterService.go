package testhelpers

import (
	"github.com/getcarina/carina/common"
	"github.com/stretchr/testify/mock"
)

type MockClusterService struct {
	common.ClusterService
	mock.Mock
}

func (mock *MockClusterService) ListClusterTemplates() ([]common.ClusterTemplate, error) {
	args := mock.Called()
	return args.Get(0).([]common.ClusterTemplate), nil
}

type StubClusterTemplate struct {
	Name     string
	COE      string
	HostType string
}

func (stub *StubClusterTemplate) GetName() string {
	return stub.Name
}

func (stub *StubClusterTemplate) GetCOE() string {
	return stub.COE
}

func (stub *StubClusterTemplate) GetHostType() string {
	return stub.HostType
}
