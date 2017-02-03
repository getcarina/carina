package client_test

import (
	"testing"

	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/internal/clienttest"
	"github.com/getcarina/carina/internal/commontest"
	"github.com/stretchr/testify/assert"
)

func TestFilterTemplatesByName(t *testing.T) {

	service := new(commontest.MockClusterService)
	service.On("ListClusterTemplates").Return([]common.ClusterTemplate{
		&commontest.StubClusterTemplate{Name: "Kubernetes 1.4.5 on LXC"},
		&commontest.StubClusterTemplate{Name: "Swarm 1.11.2 on LXC"},
	})
	account := new(clienttest.MockAccount)
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

	service := new(commontest.MockClusterService)
	service.On("ListClusterTemplates").Return([]common.ClusterTemplate{
		&commontest.StubClusterTemplate{Name: "LOUD NOISES"},
	})
	account := new(clienttest.MockAccount)
	account.On("NewClusterService").Return(service, nil)

	client := client.NewClient(false)
	templates, err := client.ListClusterTemplates(account, "*noises")
	if err != nil {
		t.Error(err)
		return
	}

	assert.Len(t, templates, 1)
}
