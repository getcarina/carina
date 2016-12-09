package makecoe

import (
	"net/http"
	"testing"

	"fmt"
	"net/http/httptest"

	"github.com/getcarina/carina/common"
)

func TestDeleteClusterThenErrorOutHaltsWait(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.RequestURI {
		case "/":
			fmt.Fprintln(w, `{"versions": [{"status": "current", "min_version": "1.0", "max_version": "1.0", "id": "v1.0"}]}`)
		case "/clusters/99999999-9999-9999-9999-999999999999":
			fmt.Fprintln(w, `{ "status": "error" }`)
		case "/v2.0/tokens":
			fmt.Fprintln(w, `{"access":{"serviceCatalog":[{"endpoints":[{"tenantId":"963451","publicURL":"https:\/\/api.dfw.getcarina.com","region":"DFW"}],"name":"cloudContainer","type":"rax:container"}],"user":{"name":"fake-user","id":"fake-userid"},"token":{"expires":"3000-01-01T12:00:00Z","id":"fake-token","tenant":{"name":"fake-tenantname","id":"fake-tenantid"}}}}`)
		default:
			fmt.Fprintln(w, "unexpected request: "+r.RequestURI)
			w.WriteHeader(404)
		}
	}))
	defer mockCarina.Close()

	acct := &Account{
		AuthEndpointOverride: mockCarina.URL + "/v2.0/",
		EndpointOverride:     mockCarina.URL,
		UserName:             "fake-user",
		APIKey:               "fake-apikey",
		Region:               "DFW",
	}

	svc := &MakeCOE{Account: acct}
	cluster := newCluster()
	cluster.Name = "test"
	cluster.ID = "99999999-9999-9999-9999-999999999999"

	err := svc.WaitUntilClusterIsDeleted(cluster)
	if err == nil {
		t.Error("WaitUntilClusterIsDeleted didn't stop when the cluster was in the error state.")
	} else {
		actualError := err.Error()
		expectedError := "Unable to delete cluster, an error occured while deleting."
		if expectedError != actualError {
			t.Errorf("Expected error: %s\nActual Error: %+v", expectedError, err)

		}
	}
}
