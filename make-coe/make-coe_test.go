package makecoe

import (
	"net/http"
	"testing"
	"regexp"

	"fmt"
	"net/http/httptest"

	"github.com/getcarina/carina/common"
)

const identityAPIVersion = "/v2.0/"

var anyClusterRegexp = regexp.MustCompile("^/clusters/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

type handler func(w http.ResponseWriter, r *http.Request)

func identityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.RequestURI {
	case "/v2.0/tokens":
		fmt.Fprintln(w, `{"access":{"serviceCatalog":[{"endpoints":[{"tenantId":"963451","publicURL":"https:\/\/api.dfw.getcarina.com","region":"DFW"}],"name":"cloudContainer","type":"rax:container"}],"user":{"name":"fake-user","id":"fake-userid"},"token":{"expires":"3000-01-01T12:00:00Z","id":"fake-token","tenant":{"name":"fake-tenantname","id":"fake-tenantid"}}}}`)
	default:
		w.WriteHeader(404)
		fmt.Fprintln(w, "unexpected request: "+r.RequestURI)
	}
}

func clusterInErrorHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case anyClusterRegexp.MatchString(r.RequestURI):
			fmt.Fprintln(w, `{ "status": "error" }`)
		default:
			fmt.Fprintln(w, "unexpected request: "+r.RequestURI)
			w.WriteHeader(404)
		}
}

func createMockCarina(h handler) (*httptest.Server, *httptest.Server) {
	return httptest.NewServer(http.HandlerFunc(h)), httptest.NewServer(http.HandlerFunc(identityHandler))
}

func createMakeCOEService(identityServer *httptest.Server, carinaServer *httptest.Server) *MakeCOE {
	acct := &Account{
		AuthEndpointOverride: identityServer.URL + identityAPIVersion,
		EndpointOverride:     carinaServer.URL,
		UserName:             "fake-user",
		APIKey:               "fake-apikey",
		Region:               "DFW",
	}

	return &MakeCOE{Account: acct}
}

func TestDeleteClusterThenErrorOutHaltsWait(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(clusterInErrorHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)
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
