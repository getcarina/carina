package makecoe

import (
	"net/http"
	"testing"
	"regexp"
	"strings"

	"fmt"
	"net/http/httptest"

	"github.com/getcarina/carina/common"
)

const identityAPIVersion = "/v2.0/"

const microversionUnsupportedJSON = `{"errors":[{"code":"make-coe-api.microverion-unsupported","detail":"If the api-version header is sent, it must be in the format 'rax:container X.Y' where 1.0 <= X.Y <= 1.0","links":[{"href":"https://getcarina.com/docs/","rel":"help"}],"max_version":"1.6","min_version":"1.5","request_id":"620c8d81-b8f9-4bb0-952b-6d08ae42eda0","status":406,"title":"Microversion unsupported"}]}`

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

func microversionUnsupportedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(406)
	fmt.Fprintln(w, microversionUnsupportedJSON)
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

func assertMicroversionUnsupportedMessaging(t *testing.T, err error) {
	microversionUnsupportedUpdateClientSubstring := "Unable to communicate with the Carina API because the client is out-of-date. Update the carina client to the latest version. See https://getcarina.com/docs/tutorials/carina-cli#update for instructions."
	microversionUnsupportedErrorMessageSubstring := "Message: Microversion unsupported - The client supports 1.0 while the server supports 1.5 - 1.6."

	actualError := err.Error()

	if ! strings.Contains(actualError, microversionUnsupportedUpdateClientSubstring) {
		t.Errorf("\nExpected error with:\n\"%s\",\nInstead got:\n\"%s\"\n", microversionUnsupportedUpdateClientSubstring, err)
	}

	if ! strings.Contains(actualError, microversionUnsupportedErrorMessageSubstring) {
		t.Errorf("\nExpected error with:\n\"%s\",\nInstead got:\n\"%s\"\n", microversionUnsupportedErrorMessageSubstring, err)
	}
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

func TestMicroversionUnsupportedListClusters(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	clusters, err := svc.ListClusters()
	if err == nil {
		t.Error("ListClusters expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if clusters != nil {
		t.Error("Expected clusters to be nil, got: ", clusters)
	}
}

func TestMicroversionUnsupportedGetCluster(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	cluster, err := svc.GetCluster("99999999-9999-9999-9999-999999999999")
	if err == nil {
		t.Error("GetCluster expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if cluster != nil {
		t.Error("Expected cluster to be nil, got: ", cluster)
	}
}

func TestMicroversionUnsupportedGetClusterCredentials(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	credentials, err := svc.GetClusterCredentials("99999999-9999-9999-9999-999999999999")
	if err == nil {
		t.Error("GetClusterCredentials expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if credentials != nil {
		t.Error("Expected credentials to be nil, got: ", credentials)
	}
}

func TestMicroversionUnsupportedResizeCluster(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	cluster, err := svc.ResizeCluster("99999999-9999-9999-9999-999999999999", 2)
	if err == nil {
		t.Error("ResizeCluster expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if cluster != nil {
		t.Error("Expected cluster to be nil, got: ", cluster)
	}
}

func TestMicroversionUnsupportedDeleteCluster(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	cluster, err := svc.DeleteCluster("99999999-9999-9999-9999-999999999999")
	if err == nil {
		t.Error("DeleteCluster expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if cluster != nil {
		t.Error("Expected cluster to be nil, got: ", cluster)
	}
}

func TestMicroversionUnsupportedCreateCluster(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	mockCarina, mockIdentity := createMockCarina(microversionUnsupportedHandler)
	defer mockCarina.Close()
	defer mockIdentity.Close()

	svc := createMakeCOEService(mockIdentity, mockCarina)

	cluster, err := svc.CreateCluster("test-cluster", "test-template", 3)
	if err == nil {
		t.Error("CreateCluster expected to return error")
	} else {
		assertMicroversionUnsupportedMessaging(t, err)
	}
	if cluster != nil {
		t.Error("Expected cluster to be nil, got: ", cluster)
	}
}
