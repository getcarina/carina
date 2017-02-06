package client_test

import (
	"testing"

	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestFilterTemplatesByName(t *testing.T) {

	service := new(testhelpers.MockClusterService)
	service.On("ListClusterTemplates").Return([]common.ClusterTemplate{
		&testhelpers.StubClusterTemplate{Name: "Kubernetes 1.4.5 on LXC"},
		&testhelpers.StubClusterTemplate{Name: "Swarm 1.11.2 on LXC"},
	})
	account := new(testhelpers.MockAccount)
	account.On("NewClusterService").Return(service, nil)

	client := client.NewClient(false)
	templates, err := client.ListClusterTemplates(account, "Kubernetes*")
	if err != nil {
		t.Error(err)
		return
	}

	assert.Len(t, templates, 1)
}

func TestFilterTemplatesByNameIsCaseInsensitive(t *testing.T) {

	service := new(testhelpers.MockClusterService)
	service.On("ListClusterTemplates").Return([]common.ClusterTemplate{
		&testhelpers.StubClusterTemplate{Name: "LOUD NOISES"},
	})
	account := new(testhelpers.MockAccount)
	account.On("NewClusterService").Return(service, nil)

	client := client.NewClient(false)
	templates, err := client.ListClusterTemplates(account, "*noises")
	if err != nil {
		t.Error(err)
		return
	}

	assert.Len(t, templates, 1)
}
