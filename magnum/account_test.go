package magnum

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getcarina/carina/common"
	"github.com/stretchr/testify/assert"
)

const identityAPIVersion = "/v3/"

type handler func(w http.ResponseWriter, r *http.Request)

func identityHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.RequestURI {
	case identityAPIVersion + "/tokens":
		fmt.Fprintln(w, `{"access":{"serviceCatalog":[{"endpoints":[{"tenantId":"963451","publicURL":"https:\/\/example.com:9511","region":"RegionOne"}],"name":"cloudContainer","type":"container-infra"}],"user":{"name":"fake-user","id":"fake-userid"},"token":{"expires":"3000-01-01T12:00:00Z","id":"fake-token","tenant":{"name":"fake-tenantname","id":"fake-tenantid"}}}}`)
	default:
		w.WriteHeader(404)
		fmt.Fprintln(w, "unexpected request: "+r.RequestURI)
	}
}

func createMockCarina(h handler) (*httptest.Server, *httptest.Server) {
	return httptest.NewServer(http.HandlerFunc(h)), httptest.NewServer(http.HandlerFunc(identityHandler))
}

func createMagnumService(identityServer *httptest.Server, carinaServer *httptest.Server) *Magnum {
	acct := &Account{
		AuthEndpoint:     identityServer.URL + identityAPIVersion,
		EndpointOverride: carinaServer.URL,
		UserName:         "fake-user",
		Password:         "fake-password",
		Domain:           "Default",
		Project:          "Default",
		Region:           "RegionOne",
	}

	return &Magnum{Account: acct}
}

func TestGopherCloudIdentityV3ErrorHandlingWorkaround(t *testing.T) {
	common.Log.RegisterTestLogger(t)

	acct := &Account{
		AuthEndpoint:     "bork://example.com" + identityAPIVersion,
		UserName:         "fake-user",
		Password:         "fake-password",
		Domain:           "Default",
		Project:          "Default",
		Region:           "RegionOne",
	}

	svc := &Magnum{Account: acct}

	_, err := svc.ListClusters()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}